package kit

import "fmt"

func blkioPathRate(path string, rate uint64, setting string) (blkioRate, error) {
	if path == "" {
		return blkioRate{}, cfgErr("", "path", setting, fmt.Errorf("%w: device path", errEmpty))
	}
	return blkioRate{path: path, rate: rate}, nil
}

// BlkioWeightDevice adds a per-device blkio weight.
func BlkioWeightDevice(path string, weight uint16) SpecOpt {
	return specFn(func(s *spec) error {
		if path == "" {
			return cfgErr("", "blkio_weight_device", "BlkioWeightDevice", fmt.Errorf("%w: device path", errEmpty))
		}
		s.blkioWDev = append(s.blkioWDev, blkioWeight{path: path, weight: weight})
		return nil
	})
}

// BlkioReadBps throttles read bytes/sec on a device.
func BlkioReadBps(path string, rate uint64) SpecOpt {
	return specFn(func(s *spec) error {
		r, err := blkioPathRate(path, rate, "BlkioReadBps")
		if err != nil {
			return err
		}
		s.blkioReadBps = append(s.blkioReadBps, r)
		return nil
	})
}

// BlkioWriteBps throttles write bytes/sec on a device.
func BlkioWriteBps(path string, rate uint64) SpecOpt {
	return specFn(func(s *spec) error {
		r, err := blkioPathRate(path, rate, "BlkioWriteBps")
		if err != nil {
			return err
		}
		s.blkioWriteBps = append(s.blkioWriteBps, r)
		return nil
	})
}

// BlkioReadIOps throttles read IOPS on a device.
func BlkioReadIOps(path string, rate uint64) SpecOpt {
	return specFn(func(s *spec) error {
		r, err := blkioPathRate(path, rate, "BlkioReadIOps")
		if err != nil {
			return err
		}
		s.blkioReadIOps = append(s.blkioReadIOps, r)
		return nil
	})
}

// BlkioWriteIOps throttles write IOPS on a device.
func BlkioWriteIOps(path string, rate uint64) SpecOpt {
	return specFn(func(s *spec) error {
		r, err := blkioPathRate(path, rate, "BlkioWriteIOps")
		if err != nil {
			return err
		}
		s.blkioWriteIOps = append(s.blkioWriteIOps, r)
		return nil
	})
}
