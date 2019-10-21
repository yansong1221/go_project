package network

import(
	"net"
	"bufio"
	_"log"
	"sync"
)

type ReadCallback func(socketID uint32, buf []byte,err error)

type TCPConn struct{
	conn net.Conn

	roundIndex,bindIndex uint16

	readBuf,writeBuf []byte

	readFunc ReadCallback

	reader *bufio.Reader
	writer *bufio.Writer

	active bool

	sendding bool

	mtx sync.Mutex
}

func newTCPConn(bindIndex uint16) *TCPConn{
	return &TCPConn{
		bindIndex : bindIndex,
		readBuf : make([]byte, 0),
		writeBuf : make([]byte, 0),
		active : false,
		sendding : false,
	}
}

func (this *TCPConn) Attach(conn net.Conn) {
	
	this.conn = conn
	this.reader = bufio.NewReader(this.conn)
	this.writer = bufio.NewWriter(this.conn)
	this.active = true
}
func (this *TCPConn) Detach(){

	this.roundIndex++
	this.active = false
	this.reader = nil
	this.writer = nil
	this.conn = nil
	this.readFunc = nil
	this.sendding = false
}

func (this *TCPConn) IsActive() bool{
	return this.active
}

func (this *TCPConn) GetScoketID() uint32{
	return uint32(this.bindIndex) << 16 | uint32(this.roundIndex)
}
func(this *TCPConn) Recv(cb ReadCallback){

	this.readFunc = cb

	go func(){

		buf := make([]byte,1024)
		n,err := this.reader.Read(buf)
		buf = buf[0:n]
		this.readFunc(this.GetScoketID(), buf,err)

	}()	
}
func(this* TCPConn) Send(buf []byte) bool{

	if buf != nil{
		this.mtx.Lock()
		this.writeBuf = append(this.writeBuf,buf...)
		this.mtx.Unlock()
	}
	
	if this.sendding || len(this.writeBuf) == 0{
		return true
	}

	this.sendding = true

	go func(){

		this.mtx.Lock()
		tmpBuf := make([]byte,len(this.writeBuf))
		copy(tmpBuf,this.writeBuf)
		this.writeBuf = this.writeBuf[0:0]
		this.mtx.Unlock()

		this.writer.Write(tmpBuf)
		this.sendding = false

		this.Send(nil)
	}()

	return true
}

func(this *TCPConn) Close(){
	this.conn.Close()
}