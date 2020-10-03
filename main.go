package main

import ("fmt"
        "flag"
        "strings"
        "strconv"
        "os"
        //"time"
        "github.com/google/gousb"
        "encoding/binary")

const (
    PRODUCT gousb.ID = 0x0c1a
    VENDOR gousb.ID = 0x1b1c

    HARDWARE_MODE byte = 0x01
    SOFTWARE_MODE byte = 0x02)

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



// led functions
type LedController struct {
    usb *UsbController
    channel byte
    numFans byte
    ledCount byte
}

func (c* LedController) Open() (err error) {
    c.channel = 0 //corsair lighting node core has 1 channel
    c.numFans = 3
    c.ledCount = c.numFans * 8 //num of fans * num of leds per fan (true for SP120 RGB PRO fan)
    c.usb = new(UsbController)
    return c.usb.Open()
}

func (c* LedController) Close() (err error) {
    if c.usb != nil {
        if err = c.usb.Close(); err != nil {
            return err
        }
    }
    c = new(LedController)
    return nil
}

func (c *LedController) Commit() (err error) {
	_, err = c.usb.SendAndRecieve([]byte {0x33, 0xff})
	return
}

func (c *LedController) Begin() (err error) {
	_, err = c.usb.SendAndRecieve([]byte {0x34, c.channel})
	return
}

func (c *LedController) SetPortState(mode byte) (err error) {
    if mode != HARDWARE_MODE && mode != SOFTWARE_MODE {
        return fmt.Errorf("invalid mode %v", mode)
    }
	_, err = c.usb.SendAndRecieve([]byte {0x30, c.channel, mode})
	return
}

func (c *LedController) getLedStripMask() (err error) {
	_, err = c.usb.SendAndRecieve([]byte {0x30, c.channel})
	return
}

func (c *LedController) Reset() (err error) {
    /* HW mode 
    Reset
    Begin
    SetPort
    SetEffect
    */
	_, err = c.usb.SendAndRecieve([]byte {0x37, c.channel})
	return
}

func (c *LedController) SetColor(r,g,b byte) (err error) {
    /*
    byte index | description
    0x00| 0x32
    0x01| Channel (0 or 1)
    0x02| Start index
    0x03| Count
    0x04| Color Channel (0: Red, 1: Green, 2: Blue)
    0x05-end| LED channel values equal to Count
    */

    var colorChannelToValue = map[byte]byte {0:r, 1:g, 2:b}
    for colorCh, v := range colorChannelToValue {
        var msg = []byte{0x32, c.channel, 0, c.ledCount, colorCh}
        var colors = make([]byte, c.ledCount)
        for i,_ := range colors {
            colors[i] = v
        }
        msg = append(msg, colors...)
        _, err = c.usb.SendAndRecieve(msg)
    }
	return
}

func (c *LedController) SetEffect(mode,r,g,b byte) (err error) {
    /*
    0x00| 0x35
    0x01| Channel (0 or 1)
    0x02| Device index
    0x03| Type of device
    0x04| Effect Mode
    0x05| Effect Speed
    0x06| Direction
    0x07| Fixed/Random Color
    0x08| 0xFF
    0x09| Color 0 Red
    0x0A| Color 0 Green
    0x0B| Color 0 Blue
    0x0C| Color 1 Red
    0x0D| Color 1 Green
    0x0E| Color 1 Blue
    0x0F| Color 2 Red
    0x10| Color 2 Green
    0x11| Color 2 Blue

    device types
    0x0A| Corsair LED Strip
    0x0C| Corsair HD-series Fan
    0x01| Corsair SP-series Fan
    0x02| Corsair ML-series Fan

    effect mode
    byte index | description
    0x00| Rainbow Wave
    0x01| Color Shift
    0x02| Color Pulse
    0x03| Color Wave
    0x04| Static
    0x05| Temperature
    0x06| Visor
    0x07| Marquee
    0x08| Blink
    0x09| Sequential
    0x0A| Rainbow
    */

	_, err = c.usb.SendAndRecieve([]byte {0x35, c.channel, 0, c.ledCount, mode, 0 /*speed*/, 0 /*direction*/, 0, 0, r, g, b})
	return
}


func (c *UsbController) GetSoftwareId() (version string, err error) {
	resp, err := c.SendAndRecieve([]byte {0x03})
	if err != nil {
		return "", err
	}
    return fmt.Sprintf("%v.%v.%v.%v", resp[1], resp[2], resp[3], resp[4]), nil
}

func (c *UsbController) GetFirmwareId() (version string, err error) {
	resp, err := c.SendAndRecieve([]byte {0x02})
	if err != nil {
		return "", err
	}
    return fmt.Sprintf("%v.%v.%v", resp[1], resp[2], resp[3]), nil
}






// Create a new type for a list of Strings
type rgbList []byte

// Implement the flag.Value interface
func (s *rgbList) String() string {
    return fmt.Sprintf("%v", *s)
}

func (s *rgbList) Set(value string) error {
    strRgb := strings.Split(value, ",")
	*s = make([]byte, len(strRgb))
	for i,v := range strRgb {
		intVal, err := strconv.Atoi(v)
		if err != nil {
			return err
		}
		(*s)[i] = byte(intVal)
	}
    return nil
}

type CaseArgs struct {
    rgb rgbList
    portState byte
}

func parseArgs() (caseArgs CaseArgs, err error) {
	// Subcommands
    caseCommand := flag.NewFlagSet("case", flag.ExitOnError)

    // Use flag.Var to create a flag of our new flagType
    // Default value is the current value at countStringListPtr (currently a nil value)
    caseCommand.Var(&(caseArgs.rgb), "RGB", "A comma seperated list of RGB values (0-255).")
    //caseCommand.String("effect", "chars", "Set effect {static|rainbow|...}. (Required)")

    /*
    countUniquePtr := caseCommand.Bool("hardware-mode", false, "Set changes to HW mode (saves after shutdown).")
    countMetricPtr := countCommand.String("effect", "chars", "Set effect {static|rainbow|...}. (Required)")
    */
	// Verify that a subcommand has been provided
    // os.Arg[0] is the main command
    // os.Arg[1] will be the subcommand
    if len(os.Args) < 2 {
        return caseArgs, fmt.Errorf("list or count subcommand is required")
    }    

	// Switch on the subcommand
    // Parse the flags for appropriate FlagSet
    // FlagSet.Parse() requires a set of arguments to parse as input
    // os.Args[2:] will be all arguments starting after the subcommand at os.Args[1]
    switch os.Args[1] {
    case "case":
        caseCommand.Parse(os.Args[2:])
    default:
        flag.PrintDefaults()
        return caseArgs, fmt.Errorf("case|cpu command required")
    }
    // Check which subcommand was Parsed using the FlagSet.Parsed() function. Handle each case accordingly.
    // FlagSet.Parse() will evaluate to false if no flags were parsed (i.e. the user did not provide any flags)
    if caseCommand.Parsed() {
        // validate RGB values
		if len(caseArgs.rgb) != 3 {
			return caseArgs, fmt.Errorf("RGB must contain 3 comma seperated integers")
        }
        for _,val := range caseArgs.rgb {
            if val > 255 || val < 0 {
                return caseArgs, fmt.Errorf("Invaid RGB value")
            }
        }
        // assign port state
        caseArgs.portState = SOFTWARE_MODE
        if caseArgs.portState {
            caseArgs.portState = HARDWARE_MODE
        }
	}
	return 
}





func main() {
    caseArgs, err := parseArgs()
    if err != nil {
        fmt.Println("Error parsing args ", err)
        return
    }
    //fmt.Println(caseArgs)
    
	//contr := UsbController{}
	contr := LedController{}
	if err:= contr.Open(); err != nil {
		fmt.Println(err)
		return
	}
	defer contr.Close()

    contr.Reset()
    contr.Begin()
    contr.SetPortState(caseArgs.portState)
    contr.SetEffect(0x04, caseArgs.rgb[0], caseArgs.rgb[1], caseArgs.rgb[2])
    contr.Commit()
    /*
    contr.Reset()
    contr.Begin()
    contr.SetPortState(HARDWARE_MODE)
    contr.SetEffect(0,0,0)
    contr.Commit()

    time.Sleep(5 * time.Second)

    contr.Reset()
    contr.Begin()
    contr.SetPortState(HARDWARE_MODE)
    contr.SetEffect(0xff,0,0xff)
    contr.Commit()

    time.Sleep(5 * time.Second)

    contr.Reset()
    contr.Begin()
    contr.SetPortState(HARDWARE_MODE)
    contr.SetEffect(0,0,0)
    contr.Commit()
    /*
    fmt.Println(contr.usb.intf)
    fmt.Println(contr.usb.inEndpoint)
    fmt.Println(contr.usb.outEndpoint)
    fmt.Println(contr.usb.GetFirmwareId())
 
    //contr.SetPortState(HARDWARE_MODE)
    contr.SetPortState(SOFTWARE_MODE)
    contr.Begin()
    for {
        contr.SetPortState(SOFTWARE_MODE)
        contr.SetColor(122,122,122)
        for i:=0;i<60;i++ {
            contr.Commit()
        }
    }
   /* 
    contr.SetPortState(HARDWARE_MODE)
    contr.Begin()
    time.Sleep(5 * time.Second)
    contr.SetColor(0,0,0)
    for i:=0;i<20;i++ {
        contr.Commit()
    }*/
}
