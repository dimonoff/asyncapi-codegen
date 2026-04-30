package main

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestFlagsSetToCommandDefaultsAcronymsToTrue(t *testing.T) {
	var flags Flags
	cmd := &cobra.Command{Use: "test"}
	flags.SetToCommand(cmd)

	require.NoError(t, cmd.Flags().Parse([]string{}))
	require.True(t, flags.Acronyms)
}

func TestFlagsSetToCommandAllowsDisablingAcronyms(t *testing.T) {
	var flags Flags
	cmd := &cobra.Command{Use: "test"}
	flags.SetToCommand(cmd)

	require.NoError(t, cmd.Flags().Parse([]string{"--acronyms=false"}))
	require.False(t, flags.Acronyms)
}

