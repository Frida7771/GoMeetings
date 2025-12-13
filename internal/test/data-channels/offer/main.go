package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"time"

	"GoMeetings/internal/helper"

	"github.com/pion/webrtc/v3"
)

func main() {

	// create peer connection
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return
	}
	defer func() {
		if err := peerConnection.Close(); err != nil {
			log.Println(err.Error())
		}
	}()

	//create data channel
	dataChannel, err := peerConnection.CreateDataChannel("foo", nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
	dataChannel.OnOpen(func() {
		log.Println("Data channel opened")
		i := -1000
		for range time.NewTicker(time.Second * 5).C {
			if err := dataChannel.SendText("hello world" + strconv.Itoa(i)); err != nil {
				log.Println(err.Error())
			}

		}

	})

	// 3.create offer
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		return
	}

	// 4. set local description
	err = peerConnection.SetLocalDescription(offer)
	if err != nil {
		return
	}

	// 5. print offer
	println("Offer:")
	println(helper.Encode(offer))

	//6.input answer
	println("Input answer:")
	var answer webrtc.SessionDescription

	answerStr, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	helper.Decode(answerStr, &answer)

	//7. set remote description
	if err := peerConnection.SetRemoteDescription(answer); err != nil {
		panic(err)
	}
	select {}

}
