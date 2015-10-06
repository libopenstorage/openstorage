package proto

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	log "github.com/Sirupsen/logrus"
	"io"
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/libopenstorage/gossip/types"
)

const (
	CONN_HOST     = "0.0.0.0"
	CONN_PORT     = "9002"
	CONN_TYPE     = "tcp"
	HEADER_LENGTH = 12
)

type ConnObj struct {
	ip         string
	rcvHandler types.OnMessageRcv
	conn       net.Conn
	listener   net.Listener
}

func connectionString(ip string) string {
	if strings.Index(ip, ":") == -1 {
		return ip + ":" + CONN_PORT
	}
	return ip
}

func NewMessageChannel(ip string) types.MessageChannel {
	// if ip string is localhost and any port,
	c, err := net.Dial(CONN_TYPE, connectionString(ip))
	if err != nil {
		return nil
	}
	return &ConnObj{conn: c, listener: nil}
}

func NewRunnableMessageChannel(addr string,
	f types.OnMessageRcv) types.MessageChannel {
	if addr == "" {
		addr = CONN_HOST + ":" + CONN_PORT
	}
	return &ConnObj{ip: connectionString(addr), rcvHandler: f}
}

func (c *ConnObj) RunOnRcvData() {

	l, err := net.Listen(CONN_TYPE, c.ip)
	if err != nil {
		log.Println("Error listening: ", err.Error())
		os.Exit(1)
	}
	c.listener = l
	defer c.listener.Close()

	for {
		tcpConn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting: ", err)
			return
		}
		connObj := &ConnObj{ip: c.ip, conn: tcpConn,
			rcvHandler: c.rcvHandler}
		connObj.rcvHandler(connObj)
		connObj.Close()
	}
}

func (c *ConnObj) Close() {
	if c.listener != nil {
		c.listener.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *ConnObj) write(buf []byte) error {
	for len(buf) > 0 {
		l, err := c.conn.Write(buf)
		if err != nil && err != syscall.EINTR {
			log.Error("Write failed: ", err)
			return err
		}
		buf = buf[l:]
	}
	return nil

}

// sendData serializes the given object and sends
// it over the given connection. Returns nil if
// it was successful, error otherwise
func (c *ConnObj) SendData(obj interface{}) error {
	err := error(nil)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(obj)
	if err != nil {
		log.Error("Failed to serialize message", err)
		return err
	}

	var header uint64 = uint64(buf.Len())
	headerBuf := make([]byte, HEADER_LENGTH)
	binary.LittleEndian.PutUint64(headerBuf[:], header)
	// first send out the header
	err = c.write(headerBuf)
	if err != nil {
		log.Error("Writing header failed with error: ", err)
		return err
	}

	// now send the actual data
	err = c.write(buf.Bytes())
	if err != nil {
		log.Error("Writing header failed with error: ", err)
		return err
	}

	return nil
}

// rcvData receives bytes over the connection
// until it can marshal the object. msg is the
// pointer to the object which will receive the data.
// Returns nil if it was successful, error otherwise.
func (c *ConnObj) RcvData(msg interface{}) error {
	msgBuffer := new(bytes.Buffer)

	for {
		// first read the header
		var header uint64
		headerLen, err := io.CopyN(msgBuffer, c.conn, HEADER_LENGTH)
		if err != nil {
			log.Error("Error reading the header: ", err)
			return err
		}
		if headerLen != HEADER_LENGTH {
			log.Error("Error reading header, read only ", headerLen, " bytes")
			return err
		}
		header = uint64(binary.LittleEndian.Uint64(msgBuffer.Bytes()))

		// now read the data
		msgBuffer.Reset()
		_, err = io.CopyN(msgBuffer, c.conn, int64(header))
		if err != nil && err != syscall.EINTR {
			log.Error("Error reading data from peer:", err)
			return err
		}

		dec := gob.NewDecoder(msgBuffer)
		err = dec.Decode(msg)
		if err != nil {
			log.Warn("Received bad packet:", err)
			return err
		} else {
			break
		}
	}

	return nil
}
