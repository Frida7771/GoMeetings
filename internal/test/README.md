# GoMeetings Test Pages

## Video Call Test Page

### Access URL
After starting the server, visit: `http://localhost:8080/test/video-call.html`

### Usage Steps

1. **Start the Server**
   ```bash
   cd internal/server
   go run main.go
   ```

2. **Open Test Page**
   - Open `http://localhost:8080/test/video-call.html` in your browser
   - You can open multiple tabs or windows to simulate multiple users

3. **Connect to Signaling Server**
   - Enter Room ID (Room Identity): e.g., `room-123`
   - Enter User ID (User Identity): e.g., `user-1`, `user-2`, etc. (each user should use a different ID)
   - (Optional) Enter JWT Token (if authentication is required)
   - Click the "Connect" button

4. **Start Video Call**
   - After connecting successfully, click the "Start Video Call" button
   - The browser will request camera and microphone permissions, please allow
   - Your local video will be displayed on the page

5. **Multi-user Call**
   - Open the same page in other browser tabs or windows
   - Use the same Room ID but different User IDs
   - Connect and start video calls
   - All users should be able to see each other's video

### Features

- ✅ Multi-user video calling (supports multiple participants)
- ✅ Audio calling
- ✅ Video toggle control
- ✅ Audio mute control
- ✅ Real-time member list display
- ✅ Automatic connection management
- ✅ Responsive UI design

### Notes

1. **Browser Permissions**: On first use, the browser will request camera and microphone permissions, which must be allowed to use the feature
2. **HTTPS Requirement**: Some browsers (like Chrome) may restrict WebRTC functionality in non-HTTPS environments. If you encounter issues, you can:
   - Use Firefox browser for testing
   - Or configure HTTPS
3. **Network Environment**: WebRTC requires STUN servers to traverse NAT. The code has configured Google's public STUN servers
4. **Room ID**: All users must use the same Room ID to see each other
5. **User ID**: Each user must use a different User ID

### Screen Share Test Pages

The original screen sharing test pages are still available:
- `http://localhost:8080/test/screen-share-plus/offer.html`
- `http://localhost:8080/test/screen-share-plus/answer.html`

### Troubleshooting

1. **Cannot see other users' video**
   - Check if everyone is using the same Room ID
   - Check the browser console for error messages
   - Confirm all users have clicked "Start Video Call"

2. **Cannot connect to signaling server**
   - Confirm the server is running
   - Check if the WebSocket URL is correct (default `ws://127.0.0.1:8080`)

3. **Camera/microphone cannot be accessed**
   - Check browser permission settings
   - Confirm no other application is using the camera/microphone
   - Try refreshing the page to request permissions again
