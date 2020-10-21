package controller

import ("fmt"
        "github.com/google/gousb")


type UsbController struct {
    ctx *gousb.Context
    dev *gousb.Device
    intfDone func()
    intf     *gousb.Interface
    inEndpoint  *gousb.InEndpoint
    outEndpoint *gousb.OutEndpoint
}

func (c *UsbController) Open() (err error) {
    c.ctx = gousb.NewContext()
    defer func() {
        if err != nil {
            c.Close()
        }
    }()

    if c.dev, err = c.ctx.OpenDeviceWithVIDPID(VENDOR, PRODUCT); err != nil { return }
    if c.dev == nil {
        c.Close()
        return fmt.Errorf("No device found with product id: %v vendor id: %v\n", PRODUCT, VENDOR)
    }

    if err = c.dev.SetAutoDetach(true); err != nil { return }
    if c.intf, c.intfDone, err = c.dev.DefaultInterface(); err != nil { return }
    if c.inEndpoint, err = c.intf.InEndpoint(0x81); err != nil { return }
	if c.outEndpoint, err = c.intf.OutEndpoint(0x01); err != nil { return }
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

	// response = make([]byte, c.inEndpoint.Desc.MaxPacketSize)
	response = make([]byte, 16)
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

