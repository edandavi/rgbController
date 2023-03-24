package controller

import ("fmt"
        "strconv"
        "strings"
        "github.com/google/gousb")


type RGB struct {
    r,g,b byte `validate:"required,gte=0,lte=255"`
}

// Implement the flag.Value interface
func (s *RGB) String() string {
    return fmt.Sprintf("r:%v g:%v b:%v", s.r, s.g, s.b)
}

func (s *RGB) Set(value string) (err error) {
    strRgb := strings.Split(value, ",")
    // validate RGB values
    if len(strRgb) != 3 {
        return fmt.Errorf("RGB must contain 3 comma seperated integers")
    }
    //r,g,b int
    r, err := strconv.Atoi(strRgb[0])
    g, err := strconv.Atoi(strRgb[1])
    b, err := strconv.Atoi(strRgb[2])
    if err != nil { return err }
    s.r, s.b, s.g = byte(r), byte(g), byte(b)
    return nil
}


type UsbController struct {
    ctx *gousb.Context
    dev *gousb.Device
    intfDone func()
    intf     *gousb.Interface
    inEndpoint  *gousb.InEndpoint
    outEndpoint *gousb.OutEndpoint
    config UsbConfig
}

type UsbConfig struct { 
    productId gousb.ID
    vendorId gousb.ID
    inEndpoint int
    outEndpoint int
    interfaceNum int
    interfaceAlternative int
}

func (c *UsbController) Open(conf UsbConfig) (err error) {
    c.ctx = gousb.NewContext()
    defer func() {
        if err != nil {
            c.Close()
        }
    }()

    if c.dev, err = c.ctx.OpenDeviceWithVIDPID(conf.vendorId, conf.productId); err != nil { return }
    if c.dev == nil {
        c.Close()
        return fmt.Errorf("No device found with product id: %v vendor id: %v\n", conf.productId, conf.vendorId)
    }

    if err = c.dev.SetAutoDetach(true); err != nil { return }
    //if c.intf, c.intfDone, err = c.dev.DefaultInterface(); err != nil { return }
    var confNum int
    var config *gousb.Config
    if confNum, err = c.dev.ActiveConfigNum(); err != nil { return }
    if config, err = c.dev.Config(confNum); err != nil { return }
    if c.intf, err = config.Interface(conf.interfaceNum, conf.interfaceAlternative); err != nil { return }
    
    if c.inEndpoint, err = c.intf.InEndpoint(conf.inEndpoint); err != nil { return }
	if c.outEndpoint, err = c.intf.OutEndpoint(conf.outEndpoint); err != nil { return }
	return
}

func (c *UsbController) Close() (err error) {
	if c.intfDone != nil {
		c.intfDone()
	}
	if c.intf != nil {
		c.intf.Close()
	}
	if c.dev != nil {
		if err = c.dev.Close(); err != nil {
			return err
		}
	}
	if c.ctx != nil {
		if err = c.ctx.Close(); err != nil {
			return err
		}
	}
    // reset pointers
    c = new(UsbController)
    return nil
}

func (c* UsbController) SendAndRecieve(msg []byte) (response []byte, err error) {
    packetSize := c.outEndpoint.Desc.MaxPacketSize
    packet := make([]byte, packetSize)
    copy(packet, msg)

	fmt.Printf("len %v send %v\n", len(packet), packet)

	if writtenBytes, err := c.outEndpoint.Write(packet); err!= nil || writtenBytes!= packetSize {
        return nil, fmt.Errorf("written %d bytes, message: %v error: %v", writtenBytes, msg, err)
	}

	response = make([]byte, c.inEndpoint.Desc.MaxPacketSize)
	// response = make([]byte, 16)
	readBytes, err := c.inEndpoint.Read(response)
	if err != nil {
		return
	}
	if readBytes == 0 {
		return response, fmt.Errorf("response is empty")
	}

	fmt.Printf("len %v read %v\n", len(response), response)

	return response, nil
}

