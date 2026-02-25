package ide

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type languageConfig struct {
	Image        string
	Filename     string
	CompileCmd   []string
	RunCmd       []string
	NeedsCompile bool
}

var languageMap = map[string]languageConfig{
	"cpp": {
		Image:        "gcc:latest",
		Filename:     "main.cpp",
		CompileCmd:   []string{"g++", "main.cpp", "-O2", "-std=c++17", "-o", "main"},
		RunCmd:       []string{"./main"},
		NeedsCompile: true,
	},
	"java": {
		Image:        "openjdk:17",
		Filename:     "Main.java",
		CompileCmd:   []string{"javac", "Main.java"},
		RunCmd:       []string{"java", "Main"},
		NeedsCompile: true,
	},
	"python": {
		Image:    "python:3.11",
		Filename: "main.py",
		RunCmd:   []string{"python", "main.py"},
	},
	"javascript": {
		Image:    "node:18",
		Filename: "main.js",
		RunCmd:   []string{"node", "main.js"},
	},
	"js": {
		Image:    "node:18",
		Filename: "main.js",
		RunCmd:   []string{"node", "main.js"},
	},
}

type DockerEvaluator struct {
	DockerBinary string
	Timeout      time.Duration
}

func NewDockerEvaluator() *DockerEvaluator {
	return &DockerEvaluator{DockerBinary: "docker", Timeout: 5 * time.Second}
}

func (e *DockerEvaluator) Evaluate(ctx context.Context, submission Submission, testCases []CodingQuestionTestCase) (EvaluationResult, error) {
	cfg, ok := languageMap[strings.ToLower(strings.TrimSpace(submission.Language))]
	if !ok {
		return EvaluationResult{Status: "failed", Detail: "unsupported language"}, nil
	}
	if len(testCases) == 0 {
		zero := 0.0
		return EvaluationResult{Status: "completed", Score: &zero, Detail: "no test cases"}, nil
	}
	tmpDir, err := os.MkdirTemp("", "ide-eval-*")
	if err != nil {
		return EvaluationResult{}, err
	}
	defer os.RemoveAll(tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, cfg.Filename), []byte(submission.Code), 0o600); err != nil {
		return EvaluationResult{}, err
	}

	if cfg.NeedsCompile {
		if _, stderr, err := e.runInDocker(ctx, cfg.Image, tmpDir, cfg.CompileCmd, ""); err != nil {
			return EvaluationResult{Status: "failed", Detail: strings.TrimSpace(stderr)}, nil
		}
	}

	var totalWeight float64
	var earned float64
	for _, tc := range testCases {
		w := tc.Weight
		if w <= 0 {
			w = 1
		}
		totalWeight += w
		stdout, stderr, err := e.runInDocker(ctx, cfg.Image, tmpDir, cfg.RunCmd, tc.Input)
		if err != nil {
			return EvaluationResult{Status: "failed", Detail: strings.TrimSpace(stderr)}, nil
		}
		if normalizeOutput(stdout) == normalizeOutput(tc.ExpectedOutput) {
			earned += w
		}
	}
	score := 0.0
	if totalWeight > 0 {
		score = (earned / totalWeight) * 100
	}
	return EvaluationResult{Status: "completed", Score: &score}, nil
}

func (e *DockerEvaluator) runInDocker(ctx context.Context, image, mountDir string, cmdArgs []string, stdin string) (string, string, error) {
	bin := e.DockerBinary
	if bin == "" {
		bin = "docker"
	}
	timeout := e.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	args := []string{"run", "--rm", "--network=none", "--memory=128m", "--cpus=0.5", "--pids-limit=64", "--read-only", "--security-opt=no-new-privileges", "-v", mountDir + ":/workspace", "-w", "/workspace", image}
	args = append(args, cmdArgs...)
	command := exec.CommandContext(runCtx, bin, args...)
	if stdin != "" {
		command.Stdin = strings.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if runCtx.Err() == context.DeadlineExceeded {
		return stdout.String(), stderr.String(), fmt.Errorf("docker execution timed out")
	}
	if err != nil {
		return stdout.String(), stderr.String(), err
	}
	return stdout.String(), stderr.String(), nil
}

func normalizeOutput(s string) string { return strings.TrimSpace(s) }
