package controller

import ("fmt"
        "strconv"
        "strings")


var config = UsbConfig{ 
    productId: 0x0c1a,
    vendorId: 0x1b1c,
    inEndpoint: 0x81,
    outEndpoint: 0x01} 

const (
    HARDWARE_MODE byte = 0x01
    SOFTWARE_MODE byte = 0x02)


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


type LncEffect struct {
    mode, speed, direction, randomColor, numOfColors byte
    rgbs [2]RGB
}


// these function are used to set cli arguments and treat unitiliazed values as non supported
func (e* LncEffect) SetSpeed(speed byte) (err error) {
    if e.speed == 0 {
        return fmt.Errorf("Effect doesn't support speed change %v", *e)
    }
    if speed > 2 {
        return fmt.Errorf("speed values are 0,1,2 got:%v", speed)
    }
    e.speed = speed
    return nil
}

func (e* LncEffect) SetDirection(direction byte) (err error) {
    if e.direction == 0 {
        return fmt.Errorf("Effect doesn't support direction change %v", *e)
    }
    if direction > 1 {
        return fmt.Errorf("speed values are 0,1 got:%v", direction)
    }
    e.direction = direction
    return nil
}

func (e* LncEffect) SetRandomColor(randomColor byte) (err error) {
    if e.randomColor  == 0 {
        return fmt.Errorf("Effect doesn't support randomColor change %v", *e)
    }
    if randomColor > 1 {
        return fmt.Errorf("randomColor values are 0,1 got:%v", randomColor)
    }
    e.randomColor = randomColor
    return nil
}

// RGB need to be set after randomColor
func (e* LncEffect) SetRGB1(rgb RGB) (err error) {
    if e.randomColor == 1 || e.numOfColors == 0 {
        return fmt.Errorf("Effect doesn't support 1 fixes colors %v", *e)
    }
    e.rgbs[0] = rgb
    return nil
}

func (e* LncEffect) SetRGB2(rgb RGB) (err error) {
    if e.randomColor == 1 || e.numOfColors != 2 {
        return fmt.Errorf("Effect doesn't support 2 fixes colors %v", *e)
    }
    e.rgbs[1] = rgb
    return nil
}

var LncEffects = map[string]LncEffect {
    "rainbow wave": LncEffect{mode:0x00, speed:1, direction:1, randomColor:1, numOfColors: 2},
    "color shift": LncEffect{mode:0x01, speed:1, randomColor:1, numOfColors: 2},
    "color pulse": LncEffect{mode:0x02, speed:1, randomColor:1, numOfColors: 2},
    "color wave": LncEffect{mode:0x03, speed:1, direction:1, randomColor:1, numOfColors: 2},
    "static": LncEffect{mode:0x04, numOfColors: 1},
    "visor": LncEffect{mode:0x06, speed:1, randomColor:1, numOfColors: 2},
    "marquee": LncEffect{mode:0x07, speed:1, numOfColors: 1},
    "strobing": LncEffect{mode:0x08, speed:1, randomColor:1},
    "sequential": LncEffect{mode:0x09, speed:1, direction:1, randomColor:1, numOfColors:1},
    "rainbow": LncEffect{mode:0x0A, speed:1, direction:1, randomColor:1},
}


// led functions
type LncController struct {
    usb *UsbController
    channel byte
    numFans byte
    ledCount byte
}

func (c* LncController) Open() (err error) {
    c.channel = 0 //corsair lighting node core has 1 channel
    c.numFans = 3
    c.ledCount = c.numFans * 8 //num of fans * num of leds per fan (true for SP120 RGB PRO fan)
    c.usb = new(UsbController)
    return c.usb.Open(config)
}

func (c* LncController) Close() (err error) {
    if c.usb != nil {
        if err = c.usb.Close(); err != nil {
            return err
        }
    }
    c = new(LncController)
    return nil
}

func (c *LncController) Commit() (err error) {
	_, err = c.usb.SendAndRecieve([]byte {0x33, 0xff})
	return
}

func (c *LncController) Begin() (err error) {
	_, err = c.usb.SendAndRecieve([]byte {0x34, c.channel})
	return
}

func (c *LncController) SetPortState(mode byte) (err error) {
    if mode != HARDWARE_MODE && mode != SOFTWARE_MODE {
        return fmt.Errorf("invalid mode %v", mode)
    }
	_, err = c.usb.SendAndRecieve([]byte {0x30, c.channel, mode})
	return
}

func (c *LncController) getLedStripMask() (err error) {
	_, err = c.usb.SendAndRecieve([]byte {0x30, c.channel})
	return
}

func (c *LncController) Reset() (err error) {
    /* HW mode 
    Reset
    Begin
    SetPort
    SetEffect
    */
	_, err = c.usb.SendAndRecieve([]byte {0x37, c.channel})
	return
}

func (c *LncController) SetColor(r,g,b byte) (err error) {
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

func (c *LncController) SetEffect(effect LncEffect) (err error) {
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
    var c1,c2 RGB
    if effect.randomColor == 0 {
        switch effect.numOfColors {
        case 1:
            c1 = effect.rgbs[0] 
        case 2:
            c1, c2 = effect.rgbs[0], effect.rgbs[1]
        default:
            return fmt.Errorf("effect cannot use RGB values %v")
        }
    }

	_, err = c.usb.SendAndRecieve([]byte {0x35, c.channel, 0, c.ledCount,
        effect.mode, effect.speed, effect.direction, effect.randomColor, 0,
        c1.r, c1.g, c1.b, c2.r, c2.g, c2.b})
	return err
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

