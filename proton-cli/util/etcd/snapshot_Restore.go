// Copyright 2021 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package etcd

import (
	"fmt"
	"strings"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/util/etcd/snapshot"

	"go.etcd.io/etcd/pkg/v3/cobrautl"
	"go.etcd.io/etcd/server/v3/datadir"
)

func SnapshotRestoreCommandFunc(logpath string, restoreCluster string,
	restoreClusterToken string,
	restoreDataDir string,
	restoreWalDir string,
	restorePeerURLs string,
	restoreName string,
	skipHashCheck bool,
	args []string) {
	if len(args) != 1 {
		err := fmt.Errorf("snapshot restore requires exactly one argument")
		cobrautl.ExitWithError(cobrautl.ExitBadArgs, err)
	}

	dataDir := restoreDataDir
	if dataDir == "" {
		dataDir = restoreName + ".etcd"
	}

	walDir := restoreWalDir
	if walDir == "" {
		walDir = datadir.ToWalDir(dataDir)
	}

	lg := GetLoggerFileAndConsole(logpath)
	sp := snapshot.NewV3(lg)

	if err := sp.Restore(snapshot.RestoreConfig{
		SnapshotPath:        args[0],
		Name:                restoreName,
		OutputDataDir:       dataDir,
		OutputWALDir:        walDir,
		PeerURLs:            strings.Split(restorePeerURLs, ","),
		InitialCluster:      restoreCluster,
		InitialClusterToken: restoreClusterToken,
		SkipHashCheck:       skipHashCheck,
	}); err != nil {
		cobrautl.ExitWithError(cobrautl.ExitError, err)
	}
}
