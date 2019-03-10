package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/notnil/chess"

	"github.com/valyala/fasthttp"

	"github.com/thanhpk/randstr"

	"github.com/gorilla/websocket"
)

type PayloadInt struct {
	I []int `json:"i"`
}

type PayloadIntString struct {
	I []int    `json:"i"`
	S []string `json:"s"`
}

type User struct {
	Name   string
	N      int
	RoomID int
	Rating int
}

type Player struct {
	User           User
	CurrentSection string
	CurrentSeat    int
}

type KurnikBot struct {
	Connection     *websocket.Conn
	CurrentPlayer  Player
	CurrentSection string
	RoomList       map[int]Room
	SectionsList   []string
	PlayerList     map[string]User
	Game           Game
	Engine         *ChessEngine
	Running        bool
}

type Game struct {
	Mux        sync.Mutex
	Turn       int
	Chess      *chess.Game
	ColorWhite bool
}

type Seat struct {
	Player User
	Taken  bool
}

type Room struct {
	N      int
	InGame bool
	Time   string
	Seat1  Seat
	Seat2  Seat
}

func (q *KurnikBot) NewRoomObject(i []int, s []string) Room {
	r := Room{}
	r.N = i[0]
	if i[1] == 1 {
		r.InGame = true
	}

	r.Time = s[0]

	s1 := Seat{}
	if i[2] == 1 {
		s1.Taken = true
	}
	if s[1] != "" {
		s1.Player = q.PlayerList[s[1]]
	}

	s2 := Seat{}
	if i[3] == 1 {
		s2.Taken = true
	}
	if s[2] != "" {
		s2.Player = q.PlayerList[s[2]]
	}

	r.Seat1 = s1
	r.Seat2 = s2

	return r
}

func BuildLoginPayload(sessionID string) PayloadIntString {
	if sessionID == "" {
		sessionID = randstr.String(16)
	}
	p := PayloadIntString{
		[]int{1710},
		[]string{
			sessionID + "+223834712694889075||",
			"en",
			"b",
			"",
			userAgent,
			fmt.Sprintf("/%d/1", time.Now().Unix()*1000),
			"w",
			"1366x768 1",
			"ref:https://www.kurnik.pl/szachy/",
			"ver:191",
		},
	}
	return p
}

func (q *KurnikBot) SendKeepAlive() {
	p := PayloadInt{[]int{2}}
	q.SendMessage(&p)
}

func (q *KurnikBot) LeaveRoom(room int) {
	p := PayloadInt{[]int{73, room}}
	q.CurrentPlayer.User.RoomID = 0
	q.SendMessage(&p)
}

func (q *KurnikBot) StartMatch() {
	p := PayloadInt{[]int{85, q.CurrentPlayer.User.RoomID}}
	q.SendMessage(&p)
}

func (q *KurnikBot) JoinRoom(roomID int) {
	p := PayloadInt{[]int{72, roomID}}
	q.CurrentPlayer.User.RoomID = roomID
	q.SendMessage(&p)
}

func (q *KurnikBot) SendChatMessage(message string) {
	p := PayloadIntString{[]int{81, q.CurrentPlayer.User.RoomID}, []string{message}}
	q.SendMessage(&p)
}

func (q *KurnikBot) CreateRoom() {
	p := PayloadInt{[]int{71}}
	q.SendMessage(&p)
}

func (q *KurnikBot) TakeSeat(seat int) {
	p := PayloadInt{[]int{83, q.CurrentPlayer.User.RoomID, seat}}
	q.CurrentPlayer.CurrentSeat = seat
	q.SendMessage(&p)
}

func (q *KurnikBot) ConnectToWebSocketServer() {
	u := url.URL{Scheme: "wss", Host: "x.kurnik.pl:17003", Path: "/ws/"}

	headers := make(http.Header)
	headers.Set("Cookie", "kt=cckn")
	headers.Set("User-Agent", userAgent)
	headers.Set("Origin", "http://kurnik.pl/")

	var err error
	q.Connection, _, err = websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		panic(err)
	}
	q.Running = true
}

func (q *KurnikBot) Disconnect() {
	q.Connection.Close()
}

func (q *KurnikBot) StartListening() {
	for q.Running {
		var p PayloadIntString
		_, b, err := q.Connection.ReadMessage()
		if err != nil {
			panic(err)
		}
		if len(b) > 0 {
			err := json.Unmarshal(b, &p)
			if err != nil {
				panic(err)
			}
			q.HandleCommands(p)
		}
	}
}

func (q *KurnikBot) HandleCommands(p PayloadIntString) {
	switch p.I[0] {
	case 1:
		q.SendKeepAlive()
	case 18:
		q.ReceiveUsername(p)
	case 24:
		q.HandlePlayerLeave(p)
	case 25:
		q.HandlePlayerUpdate(p)
	case 27:
		q.ReceivePlayerList(p)
	case 32:
		q.ReceiveSectionsList(p)
	case 33:
		q.ReceiveRating(p)
	case 70:
		q.ReceiveRoomUpdate(p)
	case 71:
		q.ReceiveRoomList(p)
	case 73:
		q.RecieveRoomCreation(p)
	case 88:
		q.RecieveRoomSeat(p)
	case 90:
		q.RecievePossibleMoves(p)
	case 91:
		q.HandleStartGame(p)
	case 92:
		q.ReceiveMove(p)
	}
}

func (q *KurnikBot) SendBestMove(p PayloadIntString) {
	start := time.Now()

	err := q.Engine.SetFEN(q.Game.Chess.FEN())
	if err != nil {
		panic(err)
	}

	res, err := q.Engine.Depth(20)
	if err != nil {
		panic(err)
	}

	from := res.BestMove[:2]
	to := res.BestMove[2:4]

	p0 := IndexByte(from[0], x)
	p1 := IndexByte(from[1], y)
	d0 := IndexByte(to[0], x)
	d1 := IndexByte(to[1], y)

	r := ((d1*8+d0)*8+p1)*8 + p0

	if len(res.BestMove) > 4 {
		fmt.Println("promotion")
		r = (IndexByte(res.BestMove[5], promotionOptions)+1)*4096 + r
	}

	elapsed := time.Since(start)
	t := elapsed.Nanoseconds() / 100000000
	if t <= 0 {
		t = 1
	}

	sp := PayloadInt{}
	sp.I = []int{92, q.CurrentPlayer.User.RoomID, 1, r, int(t)}
	fmt.Println(sp)
	// {"i":[92,2301,1,2356,47]}
	q.SendMessage(&sp)
}

func (q *KurnikBot) RecieveRoomCreation(p PayloadIntString) {
	q.CurrentPlayer.User.RoomID = p.I[1]
}

func (q *KurnikBot) RecievePossibleMoves(p PayloadIntString) {
	q.Game.Turn = p.I[3]
	if q.Game.Turn > -1 && q.CurrentPlayer.CurrentSeat == q.Game.Turn {
		q.SendBestMove(p)
	}
}

func (q *KurnikBot) RecieveRoomSeat(p PayloadIntString) {
	q.CurrentPlayer.CurrentSeat = p.I[4]
}

func (q *KurnikBot) ReceiveMove(p PayloadIntString) {
	fmt.Println("received move", p.S[0])
	err := q.Game.Chess.MoveStr(p.S[0])
	if err != nil {
		panic(err)
	}
}

func (q *KurnikBot) HandleStartGame(p PayloadIntString) {
	q.Game.Chess = chess.NewGame(chess.UseNotation(chess.AlgebraicNotation{}))
	if len(p.I) <= 2 {
		q.Game.ColorWhite = true
	}
}

func (q *KurnikBot) ReceiveSectionsList(p PayloadIntString) {
	q.SectionsList = make([]string, 0)
	sp := strings.Split(p.S[0], "\n")
	for _, v := range sp {
		q.SectionsList = append(q.SectionsList, strings.Split(v, " ")[0])
	}
	q.CurrentSection = q.SectionsList[p.I[1]]
	fmt.Println(q.CurrentSection)
}

func (q *KurnikBot) HandlePlayerUpdate(p PayloadIntString) {
	pl := User{}
	pl.Rating = p.I[3]
	pl.RoomID = p.I[2]
	pl.N = p.I[1]

	q.PlayerList[p.S[0]] = pl
}

func (q *KurnikBot) HandlePlayerLeave(p PayloadIntString) {
	delete(q.PlayerList, p.S[0])
}

func (q *KurnikBot) ReceiveCreateRoom(p PayloadIntString) {
	q.CurrentPlayer.User.RoomID = p.I[1]
}

func (q *KurnikBot) ReceiveRoomList(p PayloadIntString) {
	q.RoomList = make(map[int]Room)

	j := 0
	for i := 3; i < len(p.I)-3; i += 4 {
		r := q.NewRoomObject(p.I[i:i+4], p.S[j:j+3])

		q.RoomList[i] = r

		j += 3
	}
}

func (q *KurnikBot) ReceiveRoomUpdate(p PayloadIntString) {
	r := q.NewRoomObject(p.I[1:5], p.S[0:3])
	q.RoomList[r.N] = r

	if r.N == q.CurrentPlayer.User.RoomID {
		if r.Seat1.Taken && r.Seat2.Taken {
			q.StartMatch()
		}
	}
}

func (q *KurnikBot) RecieveRemoveRoom(p PayloadIntString) {
	delete(q.RoomList, p.I[1])
}

func (q *KurnikBot) ReceiveUsername(p PayloadIntString) {
	q.CurrentPlayer.User.Name = p.S[0]
}

func (q *KurnikBot) ReceiveRating(p PayloadIntString) {
	q.CurrentPlayer.User.Rating = p.I[1]
}

func (q *KurnikBot) JoinSection(section string) {
	p := PayloadIntString{}
	p.I = []int{20}
	p.S = []string{"/join " + section}

	q.SendMessage(&p)
}

func (q *KurnikBot) ReceivePlayerList(p PayloadIntString) {
	q.PlayerList = make(map[string]User)

	n := 3
	for _, name := range p.S {
		player := User{}
		player.Name = name
		player.N = p.I[n]
		player.RoomID = p.I[n+1]
		player.Rating = p.I[n+2]
		q.PlayerList[name] = player
	}
}

func (q *KurnikBot) Login(login, password string) {
	p := BuildLoginPayload(GetSessionID(login, password))
	q.SendMessage(&p)
}

func (q *KurnikBot) LoginAsGuest() {
	p := BuildLoginPayload("")
	q.SendMessage(&p)
}

func (q *KurnikBot) SendMessage(v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	err = q.Connection.WriteMessage(1, b)
	if err != nil {
		panic(err)
	}
}

func GetSessionID(login, password string) string {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI("https://www.kurnik.pl/login.phtml")
	req.Header.SetMethod("POST")
	req.Header.Add("User-Agent", userAgent)
	req.SetBodyString("cc=0&username=" + login + "&pw=" + password)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	client := &fasthttp.Client{
		TLSConfig: &tls.Config{InsecureSkipVerify: true},
	}

	if err := client.Do(req, resp); err != nil {
		panic(err)
	}

	// parse cookie
	// 61 =
	// 58 :
	buf := resp.Header.PeekCookie("ksession")
	var n1, n2 int
	for i := 0; i < len(buf); i++ {
		if buf[i] == 61 {
			n1 = i
			continue
		}
		if buf[i] == 58 {
			n2 = i
			break
		}
	}
	return string(buf[n1+1 : n2])
}
