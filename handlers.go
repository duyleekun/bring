package bring

import (
	"fmt"
	"strconv"
)

type Handler = func(client *Client, args []string) error

var handlers = map[string]Handler{
	"blob": func(c *Client, args []string) error {
		idx := parseInt(args[0])
		return c.streams.append(idx, args[1])
	},

	"copy": func(c *Client, args []string) error {
		srcL := parseInt(args[0])
		srcX := parseInt(args[1])
		srcY := parseInt(args[2])
		srcWidth := parseInt(args[3])
		srcHeight := parseInt(args[4])
		mask := parseInt(args[5])
		dstL := parseInt(args[6])
		dstX := parseInt(args[7])
		dstY := parseInt(args[8])
		c.display.copy(srcL, srcX, srcY, srcWidth, srcHeight,
			dstL, dstX, dstY, byte(mask))
		return nil
	},

	"cursor": func(c *Client, args []string) error {
		cursorHotspotX := parseInt(args[0])
		cursorHotspotY := parseInt(args[1])
		srcL := parseInt(args[2])
		srcX := parseInt(args[3])
		srcY := parseInt(args[4])
		srcWidth := parseInt(args[5])
		srcHeight := parseInt(args[6])
		c.display.setCursor(cursorHotspotX, cursorHotspotY,
			srcL, srcX, srcY, srcWidth, srcHeight)
		return nil
	},

	"disconnect": func(c *Client, args []string) error {
		c.session.Terminate()
		return nil
	},

	"dispose": func(c *Client, args []string) error {
		layerIdx := parseInt(args[0])
		c.display.dispose(layerIdx)
		return nil
	},

	"end": func(c *Client, args []string) error {
		idx := parseInt(args[0])
		c.streams.end(idx)
		c.streams.delete(idx)
		return nil
	},

	"error": func(c *Client, args []string) error {
		c.logger.Warnf("Received error from server: (%s) - %s", args[1], args[0])
		return nil
	},

	"img": func(c *Client, args []string) error {
		s := c.streams.get(parseInt(args[0]))
		op := byte(parseInt(args[1]))
		layerIdx := parseInt(args[2])
		//mimetype := args[3]
		x := parseInt(args[4])
		y := parseInt(args[5])
		s.onEnd = func(s *stream) {
			c.display.draw(layerIdx, x, y, op, s)
		}
		return nil
	},

	"size": func(c *Client, args []string) error {
		layerIdx := parseInt(args[0])
		w := parseInt(args[1])
		h := parseInt(args[2])
		c.display.resize(layerIdx, w, h)
		return nil
	},

	"sync": func(c *Client, args []string) error {
		if len(c.display.tasks) > 0 {
			c.logger.Debugf("Sync received. Flushing %d tasks", len(c.display.tasks))
		} else {
			c.logger.Tracef("Sync received. Flushing %d tasks", len(c.display.tasks))
		}
		err := c.display.flush()
		if err != nil {
			c.logger.Errorf("Error flushing tasks: %s", err)
		}
		if err := c.session.Send(NewInstruction("sync", args...)); err != nil {
			c.logger.Errorf("Failed send 'sync' to server: %s", err)
			return err
		}
		return nil
	},
}

func parseInt(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		fmt.Printf("Error converting '%s' to int: %s\n", s, err)
	}
	return n
}
