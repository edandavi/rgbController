package controller

//import "fmt"



// led functions
type AmdController struct {
    usb *UsbController
    channel byte
    numFans byte
    ledCount byte
}

func (c* AmdController) Open() (err error) {
    var config = UsbConfig{ 
        productId: 0x0051,
        vendorId: 0x2516,
        inEndpoint: 0x83,
        outEndpoint: 0x04,
        interfaceNum: 1,
        interfaceAlternative: 0} 
    c.usb = new(UsbController)
    return c.usb.Open(config)
}

func (c* AmdController) Close() (err error) {
    if c.usb != nil {
        if err = c.usb.Close(); err != nil {
            return err
        }
    }
    c = new(AmdController)
    return nil
}

func (c *AmdController) Begin() (err error) {
	_, err = c.usb.SendAndRecieve([]byte {0x41, 0x80})
	return err
}

func (c *AmdController) End() (err error) {
	_, err = c.usb.SendAndRecieve([]byte {0x41})
	return err
}

func (c *AmdController) Commit() (err error) {
	_, err = c.usb.SendAndRecieve([]byte {0x51, 0x28, 0x00, 0x00, 0xE0})
	return err
}

func (c *AmdController) LedSave() (err error) {
	_, err = c.usb.SendAndRecieve([]byte {0x50, 0x55})
	return err
}

func (c *AmdController) SetEffect(effect AmdEffect) (err error) {
   /* 
    0x00| 0x51
    0x01| 0x2C
    0x02| 0x01
    0x03| 0x00
    0x04| 0x05 Effect channel to edit
    0xFF| Speed (0xFF for static)
    0x06| 0x00 Direction (0:CW, 1:CCW)
    0x07| 0x01 Mode (Fan and Logo)
    0x08| 0xFF
    0x09| 0xFF Brightness
    0x0A| 0x00 Red
    0x0B| 0xFF Green
    0x0C| 0x00 Blue
    0x0D| 0x00
    0x0E| 0x00
    0x0F| 0x00
    0x10| 0xFF
    */
    if err = effect.Valid(); err != nil {return err}
	_, err = c.usb.SendAndRecieve([]byte {0x51, 0x2C, 0x01, 0x00,
        effect.channel, effect.speed, effect.direction, effect.mode, 0xFF,
        effect.brightness, effect.rgb.r, effect.rgb.g,  effect.rgb.b, 
        0, 0, 0, 0xFF})
	return err
}

func (c *AmdController) AssignLedsToEffect(logoCh, fanCh, ringCh byte) (err error) {
    ringleds := make([]byte, 15)
    for i,_ := range ringleds {
        ringleds[i] = ringCh
    }
	_, err = c.usb.SendAndRecieve(append([]byte {0x51, 0xA0, 0x01, 0, 0, 0x03, 0, 0,
        logoCh, fanCh}, ringleds...))
	return err
}

func (c *AmdController) GetVersion() (version string, err error) {
    var resp []byte
    resp, err = c.usb.SendAndRecieve([]byte {0x12, 0x20})
    if err != nil {
        return "", err
    }
    return string([]byte{resp[8], resp[10], resp[12], resp[14], resp[16], resp[18], resp[20], resp[22]}), nil
}

