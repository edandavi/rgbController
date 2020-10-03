package controller

import ("fmt")


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
    return c.usb.Open()
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

func (c *LncController) SetEffect(mode,r,g,b byte) (err error) {
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

