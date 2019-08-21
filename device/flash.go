package device

import (
    "time"
    "fmt"
)

const (
    REG_BL_MODE     = 0
    REG_BL_VERSION  = 1
    REG_BL_ERR      = 2
    REG_BL_PROG     = 3
    REG_BL_MCUID    = 4
    REG_BL_ADDR     = 8
    REG_BL_DATA     = 12
)

const (
    PROG_BL_NONE        = 0
    PROG_BL_ERASE_PAGE  = 1
    PROG_BL_READ        = 2
    PROG_BL_WRITE       = 3
    PROG_BL_EXIT        = 4
)

const (
    APP_START_ADDR   = 0x08002000
)

func (dev *Device) FlashAddress() (uint32, error) {
    var buf [4]byte
    err := dev.ReadBytes(dev.address, REG_BL_ADDR, buf[:])
    if err != nil {
        return 0, err
    }
    return uint32(buf[0]) + (uint32(buf[1]) << 8) + (uint32(buf[2]) << 16) + (uint32(buf[3]) << 24), nil
}

func (dev *Device) FlashSetAddress(addr uint32) error {
    var buf [4]byte

    buf[0] = byte(addr)
    buf[1] = byte(addr>>8)
    buf[2] = byte(addr>>16)
    buf[3] = byte(addr>>24)

    if err := dev.WriteBytes(dev.address, REG_BL_ADDR, buf[:]); err != nil {
        return err
    }
    return nil
}

func (dev *Device) waitProg(prog byte, gap time.Duration) error {
    if err := dev.WriteByte(dev.address, REG_BL_PROG, prog); err!=nil {
        return err
    }
    for i:=0; i<3; i++ {
        time.Sleep(gap)
        r, err := dev.ReadByte(dev.address, REG_BL_PROG)
        if err!=nil {
            return err
        }
        if r==0 {
            //fmt.Printf("%d",i)
            return nil
        }
        //fmt.Printf("%d", r)
    }
    return fmt.Errorf("Timed out waiting for PROG code 0x%02x to execute.", prog)
}

func (dev *Device) FlashRead(data []byte) error {
    var buf [64]byte

    if err := dev.FlashSetAddress(APP_START_ADDR); err != nil {
       return err
    }

    fmt.Printf("Reading: [")
    defer fmt.Printf("\n")
    for pos:=0; pos<len(data); pos+=64 {
        if err := dev.waitProg(PROG_BL_READ, 1 * time.Millisecond); err!=nil {
            return err
        }
        if err := dev.ReadBytes(dev.address, REG_BL_DATA, buf[:32]); err!=nil {
            return err
        }
        if err := dev.ReadBytes(dev.address, REG_BL_DATA+32, buf[32:]); err!=nil {
            return err
        }
        copy(data[pos:],buf[:])
        fmt.Print("#")
    }
    fmt.Printf("]")
    return nil
}

func (dev *Device) flashErase(block_count int) error {
    fmt.Printf("Erasing pages: [")
    defer fmt.Print("\n")

    for pos:=0; pos<block_count; pos++ {
        addr := APP_START_ADDR + uint32(pos*1024)
        //fmt.Printf("%x ", addr)
        fmt.Print("#")
        if err := dev.FlashSetAddress(addr); err != nil {
            return err
        }
        if err := dev.waitProg(PROG_BL_ERASE_PAGE, 50 * time.Millisecond); err != nil {
            return err
        }
    }
    fmt.Print("]")
    return nil
}

func (dev *Device) flashWrite(data []byte) error {
    var buf [64]byte
    
    fmt.Printf("Writing: [")
    defer fmt.Print("\n")

    for pos:=0; pos<len(data); pos+=64 {
        copy(buf[:],data[pos:])
        if err := dev.WriteBytes(dev.address, REG_BL_DATA, buf[:32]); err!=nil {
            return err
        }
        if err := dev.WriteBytes(dev.address, REG_BL_DATA+32, buf[32:]); err!=nil {
            return err
        }
        if err := dev.waitProg(PROG_BL_WRITE, 3 * time.Millisecond); err!=nil {
            return err
        }
        fmt.Print("#")
    }
    fmt.Print("]")

    return nil
}

func (dev *Device) FlashWrite(data []byte) error {

    if err := dev.flashErase((len(data)+1023)/1024); err != nil {
        return err
    }

    if err := dev.FlashSetAddress(APP_START_ADDR); err != nil {
       return err
    }

    if err := dev.flashWrite(data); err!=nil {
        return err
    }

    data_copy := make([]byte, len(data))
    if err := dev.FlashRead(data_copy); err!=nil {
        return err
    }
    for i := 0; i<len(data); i++ {
        if data[i]!=data_copy[i] {
            return fmt.Errorf("Inconsistent flash at byte %d, expected 0x%02x, found 0x%02x", i, data[i], data_copy[i])
        }
    }
    fmt.Println("Flash content verified.")
    return nil
}

func (dev *Device) FlashExit() error {
    if err := dev.WriteByte(dev.address, REG_BL_PROG, PROG_BL_EXIT); err!=nil {
        return err
    }
    return nil
}
