package monitor

import "github.com/laukkw/kwstart/kwlog"

func makeUnboundedBuffered(sendCh chan<- Blocks, log kwlog.Logger, bufferLimitWarning int) chan<- Blocks {
	ch := make(chan Blocks)

	go func() {
		var buffer []Blocks

		for {
			if len(buffer) == 0 {
				if blocks, ok := <-ch; ok {
					buffer = append(buffer, blocks)
					if len(buffer) > bufferLimitWarning {
						log.Warnf("channel buffer holds %v > %v messages", len(buffer), bufferLimitWarning)
					}
				} else {
					close(sendCh)
					break
				}
			} else {
				select {
				case sendCh <- buffer[0]:
					buffer = buffer[1:]

				case blocks, ok := <-ch:
					if ok {
						buffer = append(buffer, blocks)
						if len(buffer) > bufferLimitWarning {
							log.Warnf("channel buffer holds %v > %v messages", len(buffer), bufferLimitWarning)
						}
					}
				}
			}
		}
	}()

	return ch
}
