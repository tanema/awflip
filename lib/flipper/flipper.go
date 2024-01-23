package flipper

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

type (
	InfoType string
	Device   struct {
		PortName  string
		Port      serial.Port
		Banner    string
		ChunkSize int
	}
	FileInfo struct {
		IsDir   bool
		IsPlain bool
		Path    string
		Size    int
	}
)

const (
	cliPrompt = ">: "
	cliEOL    = "\r\n"

	InfoDevice             InfoType = "device"
	InfoPower              InfoType = "power"
	InfoPowerDebugInfoType InfoType = "power_debug"
)

func errOnly(str string, err error) error {
	return err
}

// Open will open a connection with a flipper device. If you want it to discover
// the device, pass portname as "auto", otherwise if you want to specify the
// device you want to connect to, pass its name such as /dev/flipper
func Open(portname string) (*Device, error) {
	if portname == "auto" || portname == "" {
		details, err := resolvePort()
		if err != nil {
			return nil, err
		}
		portname = details.Name
	}
	port, err := serial.Open(portname, &serial.Mode{
		BaudRate: 115200,
		DataBits: 7,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	})
	if err != nil {
		return nil, err
	}
	port.SetReadTimeout(100 * time.Microsecond)
	device := &Device{
		PortName:  portname,
		Port:      port,
		ChunkSize: 8192,
	}
	device.Banner, _ = device.readUntilNextPrompt("")
	device.Port.ResetOutputBuffer()
	return device, err
}

// Request will send a command and will read the response and return it.
func (dvc *Device) Request(cmd string, args ...any) (string, error) {
	fullCmd := fmt.Sprintf(cmd, args...) + "\r"
	if dvc.Port == nil {
		return "", errors.New("device no longer connected")
	} else if _, err := dvc.Port.Write([]byte(fullCmd)); err != nil {
		return "", err
	} else if err := dvc.Port.Drain(); err != nil {
		return "", err
	}
	return dvc.readUntilNextPrompt(fullCmd)
}

func (dvc *Device) readUntilNextPrompt(prefix string) (string, error) {
	started := false
	buffer := []byte{}
	for {
		readData := make([]byte, 64)
		n, err := dvc.Port.Read(readData)
		if n > 0 && !started {
			started = true
		}
		buffer = append(buffer, readData[:n]...)
		if started && (err != nil || n == 0) {
			break
		}
	}
	output := string(buffer)
	output = strings.TrimLeft(output, prefix)
	output = strings.TrimLeft(output, "\n")
	output = strings.TrimRight(output, cliEOL+cliPrompt)
	checkStr := strings.ToLower(output)
	if strings.Contains(checkStr, "storage error:") || strings.Contains(checkStr, "usage:") {
		return output, errors.New(output)
	}
	return output, nil
}

func (dvc *Device) Log(out io.Writer, logLevel string) error {
	if dvc.Port == nil {
		return errors.New("device no longer connected")
	} else if _, err := dvc.Port.Write([]byte("log " + logLevel)); err != nil {
		return err
	}
	_, err := io.Copy(out, dvc.Port)
	return err
}

// Info will request info from an aspect of the device. Either device, power
// or power_debug. These will return in a map in case the info returned is in
// different formats in the future.
func (dvc *Device) Info(infoType InfoType) (map[string]string, error) {
	info, err := dvc.Request("info %v", infoType)
	if err != nil {
		return nil, err
	}
	details := map[string]string{}
	lines := strings.Split(info, cliEOL)
	for _, line := range lines {
		parts := strings.Split(line, ":")
		details[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return details, nil
}

func (dvc *Device) Uptime() (time.Duration, error) {
	resp, err := dvc.Request("uptime")
	if err != nil {
		return -1, err
	}
	return time.ParseDuration(strings.TrimLeft(resp, "Uptime: "))
}

// Backlight will set the intensity of the backlight on the device. This value
// can be between 0-255
func (dvc *Device) Backlight(i int) error {
	if i < 0 || i > 255 {
		return errors.New("intesity value should be between 0-255")
	}
	return errOnly(dvc.Request("led bl %v", i))
}

// LED will set the color of the LED on the flipper device. These values should
// be between 0-255
func (dvc *Device) LED(r, g, b int) error {
	if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 {
		return errors.New("rgb values should be between 0-255")
	} else if _, err := dvc.Request("led r %v", r); err != nil {
		return err
	} else if _, err := dvc.Request("led g %v", g); err != nil {
		return err
	} else if _, err := dvc.Request("led b %v", b); err != nil {
		return err
	}
	return nil
}

func (dvc *Device) PowerOff() error {
	return errOnly(dvc.Request("power off"))
}

func (dvc *Device) Reboot() error {
	return errOnly(dvc.Request("power reboot"))
}

func (dvc *Device) Reboot2DFU() error {
	return errOnly(dvc.Request("power reboot2dfu"))
}

func (dvc *Device) Mkdir(path string) error {
	return errOnly(dvc.Request("storage mkdir %v", path))
}

func (dvc *Device) MkdirAll(path string) error {
	if err := checkPath(path); err != nil {
		return err
	}
	pathParts := strings.Split(strings.TrimPrefix(path, "/"), string(filepath.Separator))
	for i := range pathParts {
		dirName := filepath.Join(pathParts[:i+1]...)
		info, err := dvc.Stat(dirName)
		if err == nil && info.IsDir {
			continue
		} else if info.IsPlain {
			return fmt.Errorf("%v is a plain file that already exists and is not a directory", dirName)
		} else if err := dvc.Mkdir(dirName); err != nil {
			return err
		}
	}
	return nil
}

func (dvc *Device) Format(path string) error {
	if _, err := dvc.Request("storage format %v", path); err != nil {
		return err
	}
	return errOnly(dvc.Request("y"))
}

func (dvc *Device) Rm(path string) error {
	return errOnly(dvc.Request("storage remove %v", path))
}

func (dvc *Device) Copy(from, to string) error {
	return errOnly(dvc.Request("storage copy %v %v", from, to))
}

func (dvc *Device) Rename(from, to string) error {
	return errOnly(dvc.Request("storage rename %v %v", from, to))
}

func (dvc *Device) Md5(path string) (string, error) {
	return dvc.Request("storage md5 %v", path)
}

func (dvc *Device) StoreInfo(path string) (string, error) {
	return dvc.Request("storage info %v", path)
}

func (dvc *Device) Stat(path string) (FileInfo, error) {
	resp, err := dvc.Request("storage stat %v", path)
	if err != nil {
		return FileInfo{}, nil
	}
	if strings.HasPrefix(resp, "Directory") {
		return FileInfo{
			Path:  path,
			IsDir: true,
		}, nil
	} else if strings.HasPrefix(resp, "File, ") {
		size, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(resp, "File, size: "), "b"))
		return FileInfo{
			Path:    path,
			IsPlain: true,
			Size:    size,
		}, err
	}
	return FileInfo{}, errors.New("unknown file info: " + resp)
}

func (dvc *Device) Timestamp(path string) (time.Time, error) {
	resp, err := dvc.Request("storage timestamp %v", path)
	if err != nil {
		return time.Now(), err
	}
	i, err := strconv.ParseInt(strings.TrimLeft(resp, "Timestamp "), 10, 64)
	return time.Unix(i, 0), err
}

func (dvc *Device) Write(srcPath, dstPath string) error {
	if lclInfo, err := os.Stat(srcPath); err != nil {
		return err
	} else if lclInfo.IsDir() {
		return filepath.Walk(lclInfo.Name(), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			} else if info.IsDir() {
				return nil
			}
			return dvc.Write(info.Name(), filepath.Join(dstPath, filepath.Base(info.Name())))
		})
	} else if file, err := os.Open(srcPath); err != nil {
		return err
	} else if _, err := dvc.Stat(dstPath); err == nil {
		if err := dvc.Rm(dstPath); err != nil {
			return err
		}
	} else if err := dvc.MkdirAll(dstPath); err != nil {
		return err
	} else {
		for {
			chunk := make([]byte, dvc.ChunkSize)
			if n, err := file.Read(chunk); err != nil {
				return nil
			} else if _, err := dvc.Request("storage write_chunk %v %v", dstPath, n); err != nil {
				return err
			}
		}
	}
	return nil
}

func (dvc *Device) Read(srcPath, dstPath string) error {
	if info, err := dvc.Stat(srcPath); err != nil {
		return err
	} else if info.IsDir {
		rootDir := info.Path
		return dvc.Walk(info.Path, func(info FileInfo) error {
			if info.IsDir {
				return nil
			}
			return dvc.Read(info.Path, filepath.Join(dstPath, strings.TrimPrefix(info.Path, rootDir+"/")))
		})
	} else if err := os.MkdirAll(dstPath, 0755); err != nil {
		return err
	} else if out, err := os.OpenFile(filepath.Join(dstPath, filepath.Base(srcPath)), os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644); err != nil {
		return err
	} else if resp, err := dvc.Request("storage read_chunks %v %v", srcPath, dvc.ChunkSize); err != nil {
		return err
	} else if lines := strings.Split(resp, cliEOL); len(lines) == 0 || !strings.HasPrefix(lines[0], "Size: ") {
		return errors.New("malformed response")
	} else {
		fmt.Print("reading: ", filepath.Join(dstPath, filepath.Base(srcPath)))
		readSize := 0
		for readSize < info.Size {
			byteData, err := dvc.Request("y")
			if err != nil {
				return err
			}
			byteData = strings.TrimRight(byteData, cliEOL+"Ready?")
			n, err := out.Write([]byte(byteData))
			if err != nil {
				return err
			}
			readSize += n
		}
		fmt.Println(" done")
		return out.Close()
	}
}

func (dvc *Device) Walk(dir string, fn func(FileInfo) error) error {
	items, err := dvc.Ls(dir)
	if err != nil {
		return err
	}
	dirs := []FileInfo{}
	for _, info := range items {
		if err := fn(info); err != nil {
			return err
		}
		if info.IsDir {
			dirs = append(dirs, info)
		}
	}
	for _, info := range dirs {
		if err := dvc.Walk(info.Path, fn); err != nil {
			return err
		}
	}
	return nil
}

func (dvc *Device) Ls(dir string) ([]FileInfo, error) {
	if err := checkPath(dir); err != nil {
		return nil, err
	}
	if strings.HasSuffix(dir, "/") {
		dir = strings.TrimRight(dir, "/")
	}
	resp, err := dvc.Request("storage list %v", dir)
	if err != nil {
		return nil, err
	}
	dirs := []FileInfo{}
	plain := []FileInfo{}
	for _, file := range strings.Split(resp, cliEOL) {
		path := strings.TrimSpace(file)
		if strings.HasPrefix(path, "[D]") {
			dirs = append(dirs, FileInfo{
				IsDir: true,
				Path:  filepath.Join(dir, strings.TrimPrefix(path, "[D] ")),
			})
		} else if strings.HasPrefix(path, "[F]") {
			parts := strings.Split(strings.TrimPrefix(path, "[F] "), " ")
			size, err := strconv.Atoi(strings.TrimSuffix(strings.TrimSpace(parts[1]), "b"))
			if err != nil {
				return nil, err
			}
			plain = append(plain, FileInfo{
				IsPlain: true,
				Path:    filepath.Join(dir, parts[0]),
				Size:    size,
			})
		}
	}
	return append(dirs, plain...), nil
}

func (dvc *Device) Close() error {
	if dvc.Port == nil {
		return nil
	}
	err := dvc.Port.Close()
	dvc.Port = nil
	return err
}

func resolvePort() (*enumerator.PortDetails, error) {
	flippers := []*enumerator.PortDetails{}

	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, err
	}

	for _, port := range ports {
		if strings.Contains(port.Name, "flip_") {
			flippers = append(flippers, port)
		}
	}
	if len(flippers) <= 0 {
		return nil, errors.New("no connected flippers found")
	} else if len(flippers) > 1 {
		fmt.Printf("[WARN] more than one Flipper is attached, connecting to %s, other flippers found:\n", flippers[0].Name)
		for _, flipper := range flippers[1:] {
			fmt.Printf("%s : %s\n", flipper.Name, flipper.SerialNumber)
		}
		return flippers[0], nil
	}
	return flippers[0], nil
}

func checkPath(path string) error {
	if !strings.HasPrefix(path, "/int") && !strings.HasPrefix(path, "/ext") {
		return errors.New("path needs to have the prefix /int or /ext")
	}
	return nil
}
