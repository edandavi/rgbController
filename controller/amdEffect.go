package controller

import "fmt"


type AmdEffect struct {
    channel, speed, direction, mode, brightness byte
    rgb RGB
}

func (e* AmdEffect) Valid() (err error) {
    switch e.channel {
    // case "morse":
    //     if e.speed != 0x6B {
    //         return fmt.Errorf("morse speed should be 0x6B, got %v instead", e.speed)
    //     }
    case 2: // cycle
        if err = e.validSpeed([]byte{0x72,0x68,0x64,0x62,0x61}); err != nil {return err}
        if err = e.validBrightness([]byte{0x10,0x20,0x40,0x60,0x7F}); err != nil {return err}
    case 7: // rainbow
        if err = e.validSpeed([]byte{0x72,0x68,0x64,0x62,0x61}); err != nil {return err}
        if err = e.validBrightness([]byte{0x16,0x33,0x66,0x88,0xFF}); err != nil {return err}
    case 10: // swirl
        if err = e.validSpeed([]byte{0x90,0x85,0x77,0x74,0x70}); err != nil {return err}
        if err = e.validBrightness([]byte{0x33,0x66,0x99,0xCC,0xFF}); err != nil {return err}
    default:
        if err = e.validSpeed([]byte{0x3C,0x34,0x2c,0x20,0x18}); err != nil {return err}
        if err = e.validBrightness([]byte{0x33,0x66,0x99,0xCC,0xFF}); err != nil {return err}
    }

    return nil
}

func (e* AmdEffect) validSpeed(validSpeeds []byte) error {
    if contains(validSpeeds, e.speed) == false {
        return fmt.Errorf("speed should be %v, got %v instead", validSpeeds, e.speed)
    }
    return nil
}

func (e* AmdEffect) validBrightness(validBrightness []byte) error {
    if contains(validBrightness, e.brightness) == false {
        return fmt.Errorf("brightness should be %v, got %v instead", validBrightness, e.brightness)
    }
    return nil
}

func contains(arr []byte, item byte) bool {
    for _,i := range arr {
        if i == item {
            return true
        }
    }
    return false
}

var AmdEffects = map[string]AmdEffect {
    "static": AmdEffect{channel: 0, mode: 0xFF, brightness: 0xFF, speed: 0x20},
    "breathing": AmdEffect{channel: 1, mode: 0x03},
    "color cycle": AmdEffect{channel: 2, mode: 0xFF},
    "logo": AmdEffect{channel: 5},
    "fan": AmdEffect{channel: 6},
    "rainbow": AmdEffect{channel: 7, mode: 0x05},
    "bounce": AmdEffect{channel: 8, mode: 0xFF},
    "chase": AmdEffect{channel: 9, mode: 0xC3},
    "swirl": AmdEffect{channel: 10, mode: 0x4A},
    //"morse": AmdEffect{channel: 11, mode: 0x05, speed: 0x6B},
}
