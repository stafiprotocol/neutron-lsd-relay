package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stafihub/neutron-relay-sdk/client"
	"github.com/stafiprotocol/neutron-lsd-relay/pkg/config"
	"github.com/stafiprotocol/neutron-lsd-relay/pkg/utils"
)

func importArmoredAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import-armored-account",
		Args:  cobra.ExactArgs(0),
		Short: "Import ASCII-armored account",
	}
	setImportAccountFn(cmd, true)
	return cmd
}

func importUnarmoredAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import-unarmored-account",
		Args:  cobra.ExactArgs(0),
		Short: "Import unarmored hex account",
	}
	setImportAccountFn(cmd, false)
	return cmd
}

func setImportAccountFn(cmd *cobra.Command, armored bool) {
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		basePath, err := cmd.Flags().GetString(flagBasePath)
		if err != nil {
			return err
		}

		logLevelStr, err := cmd.Flags().GetString(flagLogLevel)
		if err != nil {
			return err
		}
		logLevel, err := logrus.ParseLevel(logLevelStr)
		if err != nil {
			return err
		}
		logrus.SetLevel(logLevel)

		cfg, err := config.Load(basePath)
		if err != nil {
			return err
		}

		return generateKeyFileByPrivateKey(cfg, armored)
	}
	cmd.Flags().String(flagBasePath, defaultBasePath, "base path a directory where your config.toml resids")
	cmd.Flags().String(flagLogLevel, logrus.InfoLevel.String(), "The logging level (trace|debug|info|warn|error|fatal|panic)")
}

func generateKeyFileByPrivateKey(cfg *config.Config, armoed bool) error {
	if err := os.RemoveAll(cfg.KeystorePath); err != nil {
		return err
	}

	kr, err := keyring.New("neutron", cfg.BackendOptions, cfg.KeystorePath, os.Stdin, client.MakeEncodingConfig().Marshaler)
	if err != nil {
		return fmt.Errorf("keyring.New err: %w", err)
	}
	if armoed {
		scn := bufio.NewScanner(os.Stdin)
		fmt.Println("Enter armored private key:")
		armoedPK := ""
		{
			var lines []string
			for scn.Scan() {
				line := scn.Text()
				if len(line) == 1 {
					// Group Separator (GS ^]): ctrl-]
					if line[0] == '\x1D' {
						break
					}
				}
				lines = append(lines, line)
				if line == "-----END TENDERMINT PRIVATE KEY-----" {
					break
				}
			}

			armoedPK = strings.Join(lines, "\n")
		}
		passphrase := utils.GetPassword("Enter passphrase to the imported key:")
		if err = kr.ImportPrivKey(cfg.KeyName, armoedPK, string(passphrase)); err != nil {
			return fmt.Errorf("could not import private key keypair from given string: %s", err)
		}
	} else {
		key := utils.GetPassword("Enter private key:")
		skey := string(key)
		if skey[0:2] == "0x" {
			skey = skey[2:]
		}
		if err = kr.ImportPrivKeyHex(cfg.KeyName, skey, string(hd.Secp256k1Type)); err != nil {
			return fmt.Errorf("could not import private key keypair from given string: %s", err)
		}
	}
	return nil
}
