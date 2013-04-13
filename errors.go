package main

import (
    "fmt"
    "os"
)

func showError(err error) {
    fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
}

type GLError struct {
    code int
}

func (e *GLError) Error() string { return fmt.Sprintf("GL: code=%d", e.code) }

type GLFWError struct {
    code int
}

func (e *GLFWError) Error() string { return fmt.Sprintf("GLFW: code=%d", e.code) }
