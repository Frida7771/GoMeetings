package main

import (
	"GoMeetings/internal/helper"
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/pion/webrtc/v3"
)

func main() {
	//1.create peer connection
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := peerConnection.Close(); err != nil {
			log.Println(err.Error())
		}
	}()

	//2.on data channel
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnOpen(func() {
			log.Println("Data channel opened")
		})
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Println(string(msg.Data))
		})
	})

	//3. input offer
	println("Input offer:")
	offerStr, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	var offer webrtc.SessionDescription

	helper.Decode(offerStr, &offer)

	//4. set remote description
	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		panic(err)
	}
	//5. create answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}
	//6. set local description
	if err := peerConnection.SetLocalDescription(answer); err != nil {
		panic(err)
	}
	//7.gather complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	<-gatherComplete
	//8. print answer
	fmt.Println("Answer:")
	fmt.Println(helper.Encode(peerConnection.LocalDescription()))
	//9. select
	select {}
}
