package main

import ("fmt"
        "flag"
        "strings"
        "strconv"
        "os"
        //"time"
        "rgbController/controller")
        //"encoding/binary")








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
    hwModePtr := caseCommand.Bool("hardware-mode", false, "Set changes to HW mode (saves after shutdown).")

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
        caseArgs.portState = controller.SOFTWARE_MODE
        if *hwModePtr {
            caseArgs.portState = controller.HARDWARE_MODE
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
