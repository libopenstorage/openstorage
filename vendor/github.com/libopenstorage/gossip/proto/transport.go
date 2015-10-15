package proto

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
	"net"
	"os"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/libopenstorage/gossip/types"
)

const (
	CONN_HOST     = "0.0.0.0"
	CONN_PORT     = "9002"
	CONN_TYPE     = "tcp"
	HEADER_LENGTH = 12
)

type ConnObj struct {
	Ip         string
	rcvHandler types.OnMessageRcv
	conn       net.Conn
	listener   net.Listener
}

func transportLog(conn *ConnObj, function string, err error) *log.Entry {
	connString := "nil"
	if conn != nil {
		connString = conn.Ip
	}

	return log.WithFields(log.Fields{
		"Function": function,
		"Conn":     connString,
		"Error":    err,
	})
}

func connectionString(ip string) string {
	if strings.Index(ip, ":") == -1 {
		return ip + ":" + CONN_PORT
	}
	return ip
}

func NewMessageChannel(ip string) types.MessageChannel {
	// if ip string is localhost and any port,
	c, err := net.DialTimeout(CONN_TYPE, connectionString(ip), 2*time.Second)
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
	return &ConnObj{Ip: connectionString(addr), rcvHandler: f}
}

func (c *ConnObj) RunOnRcvData() {
	fn := "RunOnRcvData"
	l, err := net.Listen(CONN_TYPE, c.Ip)
	if err != nil {
		transportLog(c, fn, err).Error("Error listening")
		os.Exit(1)
	}
	c.listener = l
	defer c.listener.Close()

	for {
		tcpConn, err := l.Accept()
		if err != nil {
			transportLog(c, fn, err).Error("Error in Accept")
			return
		}
		connObj := &ConnObj{Ip: c.Ip, conn: tcpConn,
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
			transportLog(c, "write", err).Error("Write failed")
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
	fn := "SendData"
	err := error(nil)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(obj)
	if err != nil {
		transportLog(c, fn, err).Error("Failed to serialize message")
		return err
	}

	var header uint64 = uint64(buf.Len())
	headerBuf := make([]byte, HEADER_LENGTH)
	binary.LittleEndian.PutUint64(headerBuf[:], header)
	// first send out the header
	err = c.write(headerBuf)
	if err != nil {
		transportLog(c, fn, err).Error("Writing header failed with error")
		return err
	}

	// now send the actual data
	err = c.write(buf.Bytes())
	if err != nil {
		transportLog(c, fn, err).Error("Writing header failed with error")
		return err
	}

	return nil
}

// rcvData receives bytes over the connection
// until it can marshal the object. msg is the
// pointer to the object which will receive the data.
// Returns nil if it was successful, error otherwise.
func (c *ConnObj) RcvData(msg interface{}) error {
	fn := "RcvData"
	msgBuffer := new(bytes.Buffer)

	for {
		// first read the header
		var header uint64
		headerLen, err := io.CopyN(msgBuffer, c.conn, HEADER_LENGTH)
		if err != nil {
			transportLog(c, fn, err).Error("Error reading the header")
			return err
		}
		if headerLen != HEADER_LENGTH {
			transportLog(c, fn, err).Error("Error reading header, read only ",
				headerLen, " bytes")
			return err
		}
		header = uint64(binary.LittleEndian.Uint64(msgBuffer.Bytes()))

		// now read the data
		msgBuffer.Reset()
		_, err = io.CopyN(msgBuffer, c.conn, int64(header))
		if err != nil && err != syscall.EINTR {
			transportLog(c, fn, err).Error("Error reading data from peer")
			return err
		}

		dec := gob.NewDecoder(msgBuffer)
		err = dec.Decode(msg)
		if err != nil {
			transportLog(c, fn, err).Error("Received bad packet")
			return err
		} else {
			break
		}
	}

	return nil
}
