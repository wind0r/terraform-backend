package server

import (
	"fmt"

	"github.com/spf13/viper"

	vaultclient "github.com/nimbolus/terraform-backend/pkg/client/vault"
	"github.com/nimbolus/terraform-backend/pkg/kms"
	"github.com/nimbolus/terraform-backend/pkg/kms/local"
	"github.com/nimbolus/terraform-backend/pkg/kms/transit"
)

func GetKMS() (k kms.KMS, err error) {
	viper.SetDefault("kms_backend", "local")
	backend := viper.GetString("kms_backend")

	switch backend {
	case "local":
		key := viper.GetString("kms_key")
		if key == "" {
			key, _ = local.GenerateKey()
			return nil, fmt.Errorf("no key for local KMS defined, set KMS_KEY (e.g. to this generated key: %s)", key)
		}

		k, err = local.NewLocalKMS(key)
	case "vault":
		var key string
		keyPath := viper.GetString("kms_vault_key_path")
		if keyPath == "" {
			return nil, fmt.Errorf("no vault key path for Vault KMS defined, set KMS_VAULT_KEY_PATH")
		}

		if vaultClient, err := vaultclient.NewVaultClient(); err != nil {
			return nil, fmt.Errorf("failed to setup Vault client for Vault KMS: %v", err)
		} else if key, err = vaultclient.GetKvValue(vaultClient, keyPath, "key"); err != nil {
			return nil, fmt.Errorf("failed to get key for Vault KMS: %v", err)
		}

		k, err = local.NewLocalKMS(key)
	case "transit":
		k, err = transit.NewVaultTransit(viper.GetString("kms_transit_engine"), viper.GetString("kms_transit_key"))
	default:
		return nil, fmt.Errorf("failed to initialize KMS backend %s: %v", backend, err)
	}
	return
}
