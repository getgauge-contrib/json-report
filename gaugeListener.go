package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/apoorvam/json-report/gauge_messages"
	"github.com/golang/protobuf/proto"
)

type GaugeResultHandlerFn func(*gauge_messages.SuiteExecutionResult)

type GaugeListener struct {
	connection      net.Conn
	onResultHandler GaugeResultHandlerFn
}

func newGaugeListener(host string, port string) (*GaugeListener, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err == nil {
		return &GaugeListener{connection: conn}, nil
	}
	return nil, err
}

func (gaugeListener *GaugeListener) OnSuiteResult(resultHandler GaugeResultHandlerFn) {
	gaugeListener.onResultHandler = resultHandler
}

func (gaugeListener *GaugeListener) Start() {
	buffer := new(bytes.Buffer)
	data := make([]byte, 8192)
	for {
		n, err := gaugeListener.connection.Read(data)
		if err != nil {
			return
		}
		buffer.Write(data[0:n])
		gaugeListener.processMessages(buffer)
	}
}

func (gaugeListener *GaugeListener) processMessages(buffer *bytes.Buffer) {
	for {
		messageLength, bytesRead := proto.DecodeVarint(buffer.Bytes())
		if messageLength > 0 && messageLength < uint64(buffer.Len()) {
			message := &gauge_messages.Message{}
			messageBoundary := int(messageLength) + bytesRead
			err := proto.Unmarshal(buffer.Bytes()[bytesRead:messageBoundary], message)
			if err != nil {
				log.Printf("Failed to read proto message: %s\n", err.Error())
			} else {
				if *message.MessageType == gauge_messages.Message_KillProcessRequest {
					gaugeListener.connection.Close()
					os.Exit(0)
				}
				if *message.MessageType == gauge_messages.Message_SuiteExecutionResult {
					result := message.GetSuiteExecutionResult()
					gaugeListener.onResultHandler(result)
				}
				buffer.Next(messageBoundary)
				if buffer.Len() == 0 {
					return
				}
			}
		} else {
			return
		}
	}
}
