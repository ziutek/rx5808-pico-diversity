// Copyright 2026 Michał Derkacz. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"time"

	"github.com/embeddedgo/display/pix/displays"
	"github.com/embeddedgo/pico/dci/tftdci"
	"github.com/embeddedgo/pico/devboard/pico2/board/pins"
	"github.com/embeddedgo/pico/hal/i2c"
	"github.com/embeddedgo/pico/hal/i2c/i2c0dma"
	"github.com/embeddedgo/pico/hal/pio"
	"github.com/embeddedgo/pico/hal/system/console/uartcon"
	"github.com/embeddedgo/pico/hal/uart"
	"github.com/embeddedgo/pico/hal/uart/uart0"
)

func main() {
	// Used IO pins
	const (
		// Serial console
		conTx = pins.GP0
		conRx = pins.GP1

		// Video Receiver SPI interface (three consecutive pins)
		vrxClk  = pins.GP2
		vrxLE   = pins.GP3
		vrxData = pins.GP4

		// Display I2C interface
		dispSDA = pins.GP20
		dispSCL = pins.GP21
	)

	// Serial console
	uartcon.Setup(uart0.Driver(), conRx, conTx, uart.Word8b, 115200, "UART0")

	// I2C display
	m := i2c0dma.Master()
	m.UsePin(dispSDA, i2c.SDA)
	m.UsePin(dispSCL, i2c.SCL)
	m.Setup(400e3)
	dci := tftdci.NewI2C(m, 0b0111100)
	disp := displays.Adafruit_1i3_128x64_OLED_SH1106.New(dci)
	_ = disp

	// Video receiver
	pio0 := pio.Block(0)
	pio0.SetReset(true)
	pio0.SetReset(false)
	pio0.Load(pioProg_rtc6715spi, 0)
	sm := pio0.SM(0)
	sm.UsePin(vrxClk, pio.Out)
	sm.UsePin(vrxLE, pio.Out)
	sm.UsePin(vrxData, pio.InOut)
	sm.Configure(pioProg_rtc6715spi, 0, 0)
	sm.SetPinBase(vrxData, vrxData, vrxLE, vrxClk)
	sm.SetClkFreq(1e4)
	sm.Enable()
	printRegs(sm)

	regs := []string{
		0x0: "Synthezier Reg A        ",
		0x1: "Synthezier Reg B        ",
		0x2: "Synthezier Reg C        ",
		0x3: "Synthezier Reg D        ",
		0x4: "VCO Switch-Cap Ctrl Reg ",
		0x5: "DFC Ctrl Reg            ",
		0x6: "6M Audio Demod Ctrl Reg ",
		0x7: "6M5 Audio Demod Ctrl Reg",
		0x8: "Receiver Ctrl Reg 1     ",
		0x9: "Receiver Ctrl Reg 2     ",
		0xa: "Power Down Ctrl Reg     ",
		0xf: "State Reg               ",
	}

	for {
		for addr, name := range regs {
			if name == "" {
				continue
			}
			sm.WriteWord32(uint32(addr) << 1)
			v, _ := sm.ReadWord32()
			fmt.Printf("%#x: %s : %020b\n", addr, name, v)
		}
		fmt.Println("---")
		time.Sleep(5 * time.Second)
	}

}

func printRegs(sm *pio.SM) {
	sr := sm.Regs()
	fmt.Printf("CLKDIV:    %08x\n", sr.CLKDIV.Load())
	fmt.Printf("EXECCTRL:  %08x\n", sr.EXECCTRL.Load())
	fmt.Printf("SHIFTCTRL: %08x\n", sr.SHIFTCTRL.Load())
	fmt.Printf("ADDR:      %08x\n", sr.ADDR.Load())
	fmt.Printf("PINCTRL:   %08x\n", sr.PINCTRL.Load())
}
