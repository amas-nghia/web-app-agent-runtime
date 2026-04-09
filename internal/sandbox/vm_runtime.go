package sandbox

import (
	"context"
	"fmt"
	"os/exec"
)

type vmRuntime struct{}

func newVMRuntime() Runtime { return vmRuntime{} }

func (vmRuntime) Start(ctx context.Context, spec Spec) (Handle, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if _, err := exec.LookPath("qemu-system-x86_64"); err != nil {
		return nil, fmt.Errorf("vm runtime: qemu-system-x86_64 not found: %w", err)
	}
	return nil, fmt.Errorf("vm runtime: not implemented yet for run %s", spec.RunID)
}

var _ Runtime = vmRuntime{}
var _ Handle = (*noopHandle)(nil)
