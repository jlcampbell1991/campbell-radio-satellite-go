package player

import (
	"bufio"
	"fmt"
	"math"
	"os/exec"
	"regexp"
	"strconv"
	"syscall"

	"github.com/jlcampbell1991/campbell-radio-satellite-go/internal/clients"
	"github.com/jlcampbell1991/campbell-radio-satellite-go/internal/models"
	"github.com/jlcampbell1991/campbell-radio-satellite-go/internal/utilities"
)

type Player interface {
	Play(medium models.Media, callback func()) error

	PauseResume() error

	SetVolume(volume int) error

	GetVolume() (*int, error)
}

type player struct {
	logger       clients.Logger
	currentProc  *exec.Cmd
	paused       bool
	ffplay       string
	amixer       string
	amixerDevice string
}

func NewPlayer(ffplay, amixer, amixerDevice string, logger clients.Logger) Player {
	return &player{
		logger:       logger,
		ffplay:       ffplay,
		amixer:       amixer,
		amixerDevice: amixerDevice,
	}
}

func (p *player) Play(medium models.Media, callback func()) error {
	p.logger.Info("%+v", medium)

	threshold := medium.Loudness + medium.LRange/2

	const (
		minRange = 4.0
		maxRange = 20.0
		minRatio = 1.5
		maxRatio = 4.0
	)

	clampedLR := math.Max(minRange, math.Min(maxRange, medium.LRange))

	ratio := minRatio +
		(clampedLR-minRange)*(maxRatio-minRatio)/(maxRange-minRange)

	const attack = 15

	release := math.Round(
		150 + (clampedLR-minRange)*(800-150)/(maxRange-minRange),
	)

	compressor := fmt.Sprintf(
		"acompressor=threshold=%.1fdB:ratio=%.2f:attack=%d:release=%.0f",
		threshold,
		ratio,
		attack,
		release,
	)

	if p.currentProc != nil {
		p.currentProc.Process.Kill()
		p.currentProc = nil
		p.paused = false
	}

	audioFilter := fmt.Sprintf(
		"volume=%.2fdB,%s,alimiter=limit=0.95,apad=pad_dur=3,afade=t=out:st=%.0d:d=%d",
		medium.Gain,
		compressor,
		medium.Duration-medium.FadeOut-1,
		medium.FadeOut,
	)

	url := fmt.Sprintf(
		"http://192.168.50.185:9090/media/%s/stream",
		medium.ID,
	)

	cmd := exec.Command(
		p.ffplay,
		"-autoexit",
		"-analyzeduration", "100M",
		"-probesize", "100M",
		"-i", url,
		"-nodisp",
		"-nostats",
		"-hide_banner",
		"-af", audioFilter,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ffplay failed: %w", err)
	}

	p.currentProc = cmd
	p.paused = false

	utilities.Go(func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			p.logger.Info("%s", scanner.Text())
		}
	})

	utilities.Go(func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			p.logger.Info("ffplay: %s", scanner.Text())
		}
	})

	utilities.Go(func() {
		err := cmd.Wait()

		// A newer process may have replaced this one.
		if p.currentProc != cmd {
			return
		}

		p.currentProc = nil

		if err != nil {
			p.logger.Error("ffplay exited: %v", err)
		} else {
			p.logger.Info("ffplay exited successfully")
		}

		if callback != nil {
			callback()
		}
	})

	return nil
}

func (p *player) PauseResume() error {
	if p.currentProc == nil {
		p.logger.Error("No current ffplay process")
		return fmt.Errorf("no current ffplay process")
	}

	if p.paused {
		if err := p.currentProc.Process.Signal(syscall.SIGCONT); err != nil {
			return fmt.Errorf("failed to resume ffplay: %w", err)
		}

		p.paused = false
		p.logger.Info("ffplay resumed")
	} else {
		if err := p.currentProc.Process.Signal(syscall.SIGSTOP); err != nil {
			return fmt.Errorf("failed to pause ffplay: %w", err)
		}

		p.paused = true
		p.logger.Info("ffplay paused")
	}

	return nil
}

func (p *player) SetVolume(volume int) error {
	cmd := exec.Command(
		p.amixer,
		"set",
		p.amixerDevice,
		fmt.Sprintf("%d%%", volume),
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf(
			"failed to set volume: %w: %s",
			err,
			output,
		)
	}

	p.logger.Info("amixer volume set to %d", volume)

	return nil
}

func (p *player) GetVolume() (*int, error) {
	cmd := exec.Command(
		p.amixer,
		"get",
		p.amixerDevice,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get volume: %w", err)
	}

	re := regexp.MustCompile(`\[(\d+)%\]`)
	match := re.FindSubmatch(output)

	if len(match) < 2 {
		return nil, fmt.Errorf("could not find volume in amixer output")
	}

	volume, err := strconv.Atoi(string(match[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid volume: %w", err)
	}

	p.logger.Info("amixer volume: %d", volume)

	return &volume, nil
}
