package main

import ("fmt"
        "flag"
        "strings"
        "os"
        //"time"
        "gopkg.in/go-playground/validator.v9"
        "rgbController/controller")
        //"encoding/binary")



type CaseArgs struct {
    portState byte
    effect controller.LncEffect
}

func parseArgs() (caseArgs CaseArgs, err error) {
	// Subcommands
    caseCommand := flag.NewFlagSet("case", flag.ExitOnError)
    

    var rgb1,rgb2 controller.RGB

    // Use flag.Var to create a flag of our new flagType
    // Default value is the current value at countStringListPtr (currently a nil value)
    caseCommand.Var(&rgb1, "RGB1", "A comma seperated list of RGB values (0-255).")
    caseCommand.Var(&rgb2, "RGB2", "A comma seperated list of RGB values (0-255).")
    hwModePtr := caseCommand.Bool("hardware-mode", false, "Set changes to HW mode (saves after shutdown).")
    effectPtr := caseCommand.String("effect", "chars",
    "Set effect {Rainbow Wave|Color Shift|Color Pulse|Color Wave|Static|Visor|Marquee|strobing|Sequential|Rainbow}. (Required)")
    

    /*
    countUniquePtr := caseCommand.Bool("hardware-mode", false, "Set changes to HW mode (saves after shutdown).")
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
        var ok bool
        caseArgs.effect, ok = controller.LncEffects[strings.ToLower(*effectPtr)]
        if !ok {
            return caseArgs, fmt.Errorf("%v is not a supported effect", *effectPtr)
        }
        caseArgs.effect.Rgbs[0] = rgb1
        caseArgs.effect.Rgbs[1] = rgb2

        caseArgs.portState = controller.SOFTWARE_MODE
        if *hwModePtr {
            caseArgs.portState = controller.HARDWARE_MODE
        }

        validate := validator.New()
        errs := validate.Struct(&caseArgs)
        if errs != nil {
            return caseArgs, errs
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
    contr := controller.LncController{}
	if err:= contr.Open(); err != nil {
		fmt.Println(err)
		return
	}
	defer contr.Close()

    fmt.Println("effect: %v", caseArgs.effect)

    contr.Reset()
    contr.Begin()
    contr.SetPortState(caseArgs.portState)
    contr.SetEffect(caseArgs.effect)
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
