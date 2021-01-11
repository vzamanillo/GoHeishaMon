package main

import (
	"encoding/hex"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/xid"
)

func makeSwitchTopic(name string, state string) {
	var t autoDiscoverStruct
	t.Name = fmt.Sprintf("TEST-%s", name)
	t.StateTopic = config.MqttTopicBase + "/" + state
	t.CommandTopic = config.MqttSetBase + "/" + name
	t.UID = fmt.Sprintf("Aquarea-%s-%s", config.MqttLogin, t.Name)
	switchTopics[name] = t

}

func startsub(c mqtt.Client) {
	var t autoDiscoverStruct
	c.Subscribe(config.MqttSetBase+"/SetHeatpump", 2, handleSetHeatpump)
	makeSwitchTopic("SetHeatpump", "Heatpump_State")
	c.Subscribe(config.MqttSetBase+"/SetQuietMode", 2, handleSetQuietMode)
	c.Subscribe(config.MqttSetBase+"/SetZ1HeatRequestTemperature", 2, handleSetZ1HeatRequestTemperature)
	c.Subscribe(config.MqttSetBase+"/SetZ1CoolRequestTemperature", 2, handleSetZ1CoolRequestTemperature)
	c.Subscribe(config.MqttSetBase+"/SetZ2HeatRequestTemperature", 2, handleSetZ2HeatRequestTemperature)
	c.Subscribe(config.MqttSetBase+"/SetZ2CoolRequestTemperature", 2, handleSetZ2CoolRequestTemperature)
	c.Subscribe(config.MqttSetBase+"/SetOperationMode", 2, handleSetOperationMode)
	c.Subscribe(config.MqttSetBase+"/SetForceDHW", 2, handleSetForceDHW)
	makeSwitchTopic("SetForceDHW", "Force_DHW_State")
	c.Subscribe(config.MqttSetBase+"/SetForceDefrost", 2, handleSetForceDefrost)
	makeSwitchTopic("SetForceDefrost", "Defrosting_State")
	c.Subscribe(config.MqttSetBase+"/SetForceSterilization", 2, handleSetForceSterilization)
	makeSwitchTopic("SetForceSterilization", "Sterilization_State")
	c.Subscribe(config.MqttSetBase+"/SetHolidayMode", 2, handleSetHolidayMode)
	makeSwitchTopic("SetHolidayMode", "Holiday_Mode_State")
	c.Subscribe(config.MqttSetBase+"/SetPowerfulMode", 2, handleSetPowerfulMode)

	t.Name = "TEST-SetPowerfulMode-30min"
	t.CommandTopic = config.MqttSetBase + "/SetPowerfulMode"
	t.StateTopic = config.MqttTopicBase + "/Powerful_Mode_Time"
	t.UID = fmt.Sprintf("Aquarea-%s-%s", config.MqttLogin, t.Name)
	t.PayloadOn = "1"
	t.StateON = "on"
	t.StateOff = "off"
	t.ValueTemplate = `{%- if value == "1" -%} on {%- else -%} off {%- endif -%}`
	switchTopics["SetPowerfulMode1"] = t
	t = autoDiscoverStruct{}
	t.Name = "TEST-SetPowerfulMode-60min"
	t.CommandTopic = config.MqttSetBase + "/SetPowerfulMode"
	t.StateTopic = config.MqttTopicBase + "/Powerful_Mode_Time"
	t.UID = fmt.Sprintf("Aquarea-%s-%s", config.MqttLogin, t.Name)
	t.PayloadOn = "2"
	t.StateON = "on"
	t.StateOff = "off"
	t.ValueTemplate = `{%- if value == "2" -%} on {%- else -%} off {%- endif -%}`
	switchTopics["SetPowerfulMode2"] = t
	t = autoDiscoverStruct{}
	t.Name = "TEST-SetPowerfulMode-90min"
	t.CommandTopic = config.MqttSetBase + "/SetPowerfulMode"
	t.StateTopic = config.MqttTopicBase + "/Powerful_Mode_Time"
	t.UID = fmt.Sprintf("Aquarea-%s-%s", config.MqttLogin, t.Name)
	t.PayloadOn = "3"
	t.StateON = "on"
	t.StateOff = "off"
	t.ValueTemplate = `{%- if value == "3" -%} on {%- else -%} off {%- endif -%}`
	switchTopics["SetPowerfulMode3"] = t
	t = autoDiscoverStruct{}

	c.Subscribe(config.MqttSetBase+"/SetDHWTemp", 2, handleSetDHWTemp)
	c.Subscribe(config.MqttSetBase+"/SendRawValue", 2, handleSendRawValue)
	if config.EnableCommand == true {
		c.Subscribe(config.MqttSetBase+"/OSCommand", 2, handleOSCommand)
	}

	//Perform additional action...
}

func handleOSCommand(mclient mqtt.Client, msg mqtt.Message) {
	var cmd *exec.Cmd
	var out2 string
	s := strings.Split(string(msg.Payload()), " ")
	if len(s) < 2 {
		cmd = exec.Command(s[0])
	} else {
		cmd = exec.Command(s[0], s[1:]...)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		// TODO: handle error more gracefully
		out2 = fmt.Sprintf("%s", err)
	}
	comout := fmt.Sprintf("%s - %s", out, out2)
	TOP := fmt.Sprintf("%s/OSCommand/out", config.MqttSetBase)
	fmt.Println("Publikuje do ", TOP, "warosc", string(comout))
	token := mclient.Publish(TOP, byte(0), false, comout)
	if token.Wait() && token.Error() != nil {
		fmt.Printf("Fail to publish, %v", token.Error())
	}

}

func handleSendRawValue(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	cts := strings.TrimSpace(string(msg.Payload()))
	command, err = hex.DecodeString(cts)
	if err != nil {
		fmt.Println(err)
	}

	commandsToSend[xid.New()] = command
}

func handleSetOperationMode(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var setMode byte
	a, _ := strconv.Atoi(string(msg.Payload()))

	switch a {
	case 0:
		setMode = 82
	case 1:
		setMode = 83
	case 2:
		setMode = 89
	case 3:
		setMode = 33
	case 4:
		setMode = 98
	case 5:
		setMode = 99
	case 6:
		setMode = 104
	default:
		setMode = 0
	}

	fmt.Printf("set heat pump mode to  %d", setMode)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, 0x00, 0x00, setMode, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	commandsToSend[xid.New()] = command
}

func handleSetDHWTemp(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var heatpumpState byte

	a, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		a = int(f)
	}

	e := a + 128
	heatpumpState = byte(e)
	fmt.Printf("set DHW temperature to   %d", a)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, heatpumpState, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	commandsToSend[xid.New()] = command
}

func handleSetPowerfulMode(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var heatpumpState byte

	a, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		a = int(f)
	}

	e := a + 73
	heatpumpState = byte(e)
	fmt.Printf("set powerful mode to  %d", a)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, 0x00, 0x00, 0x00, heatpumpState, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	commandsToSend[xid.New()] = command
}

func handleSetHolidayMode(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var heatpumpState byte
	e := 16
	a, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		a = int(f)
	}

	if a == 1 {
		e = 32
	}
	heatpumpState = byte(e)
	fmt.Printf("set holiday mode to  %d", heatpumpState)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, 0x00, heatpumpState, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	commandsToSend[xid.New()] = command
}

func handleSetForceSterilization(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var heatpumpState byte
	e := 0
	a, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		a = int(f)
	}

	if a == 1 {
		e = 4
	}
	heatpumpState = byte(e)
	fmt.Printf("set force sterilization  mode to %d", heatpumpState)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, 0x00, 0x00, 0x00, 0x00, heatpumpState, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	commandsToSend[xid.New()] = command
}

func handleSetForceDefrost(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var heatpumpState byte
	e := 0
	a, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		a = int(f)
	}

	if a == 1 {
		e = 2
	}
	heatpumpState = byte(e)
	fmt.Printf("set force defrost mode to %d", heatpumpState)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, 0x00, 0x00, 0x00, 0x00, heatpumpState, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	commandsToSend[xid.New()] = command
}

func handleSetForceDHW(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var heatpumpState byte
	e := 64
	a, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		a = int(f)
	}

	if a == 1 {
		e = 128
	}
	heatpumpState = byte(e)
	fmt.Printf("set force DHW mode to %d", heatpumpState)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, heatpumpState, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	commandsToSend[xid.New()] = command
}

func handleSetZ1HeatRequestTemperature(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var requestTemp byte
	e, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		e = int(f)
	}

	e = e + 128
	requestTemp = byte(e)
	fmt.Printf("set z1 heat request temperature to %d", requestTemp-128)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, requestTemp, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	commandsToSend[xid.New()] = command
}

func handleSetZ1CoolRequestTemperature(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var requestTemp byte
	e, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		e = int(f)
	}
	e = e + 128
	requestTemp = byte(e)
	fmt.Printf("set z1 cool request temperature to %d", requestTemp-128)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, requestTemp, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	commandsToSend[xid.New()] = command
}

func handleSetZ2HeatRequestTemperature(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var requestTemp byte
	e, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		e = int(f)
	}
	e = e + 128
	requestTemp = byte(e)
	fmt.Printf("set z2 heat request temperature to %d", requestTemp-128)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, requestTemp, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	commandsToSend[xid.New()] = command
}

func handleSetZ2CoolRequestTemperature(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var requestTemp byte
	e, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		e = int(f)
	}
	e = e + 128
	requestTemp = byte(e)
	fmt.Printf("set z2 cool request temperature to %d", requestTemp-128)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, requestTemp, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	commandsToSend[xid.New()] = command
}

func handleSetQuietMode(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var quietMode byte

	e, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		e = int(f)
	}
	e = (e + 1) * 8

	quietMode = byte(e)
	fmt.Printf("set Quiet mode to %d", quietMode/8-1)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, 0x00, 0x00, 0x00, quietMode, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	commandsToSend[xid.New()] = command
}

func handleSetHeatpump(mclient mqtt.Client, msg mqtt.Message) {
	var command []byte
	var heatpumpState byte

	e := 1
	a, er := strconv.Atoi(string(msg.Payload()))
	if er != nil {
		f, _ := strconv.ParseFloat(string(msg.Payload()), 64)
		a = int(f)
	}
	if a == 1 {
		e = 2
	}

	heatpumpState = byte(e)
	fmt.Printf("set heatpump state to %d", heatpumpState)
	command = []byte{0xf1, 0x6c, 0x01, 0x10, heatpumpState, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if config.Loghex == true {
		logHex(command, len(command))
	}
	commandsToSend[xid.New()] = command
}
