// Copyright (C) 2023  Tricorder Observability
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package module

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/tricorder/src/cli/pkg/output"

	apiserver "github.com/tricorder/src/api-server/http"
	"github.com/tricorder/src/utils/file"
	"github.com/tricorder/src/utils/log"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an eBPF+WASM module",
	Long: "Create an eBPF+WASM module with BCC source file and WASM binary file. For example:\n" +
		"$ starship-cli module create --api-server=<address> -m <module_json_file> -b <bcc_source_file> " +
		"-w <wasm_binary_file>",
	Run: func(cmd *cobra.Command, args []string) {
		bccStr, err := file.Read(bccFilePath)
		if err != nil {
			log.Fatalf("Failed to read --bcc-file-path='%s', error: %v", bccFilePath, err)
		}

		wasmBytes, err := file.ReadBin(wasmFilePath)
		if err != nil {
			log.Fatalf("Failed to read --wasm-file-path='%s', error: %v", wasmFilePath, err)
		}

		moduleReq, err := parseModuleJsonFile(moduleFilePath)
		if err != nil {
			log.Fatalf("Failed to read --module-json-path='%s', error: %v", moduleFilePath, err)
		}
		// override bcc code contet by bcc file
		moduleReq.Ebpf.Code = bccStr
		// override wasm code contet by wasm file
		moduleReq.Wasm.Code = wasmBytes
		client := apiserver.NewClient(apiServerAddress)
		resp, err := client.CreateModule(moduleReq)
		if err != nil {
			log.Error(err)
			return
		}

		// TODO(jun): refactor output to delete this hack
		// we can upgrade golang version and introduce generic code
		// to provide a generic interface to output
		respByte, err := json.Marshal(resp)
		if err != nil {
			log.Error(err)
			return
		}

		err = output.Print(outputFormat, respByte)
		if err != nil {
			log.Error(err)
		}
	},
}

// the file path of module in json format flag
var (
	moduleFilePath string
	bccFilePath    string
	wasmFilePath   string
)

func init() {
	createCmd.Flags().StringVarP(&moduleFilePath, "module", "m",
		moduleFilePath, "The path of the JSON file that describes an eBPF+WASM module.")
	createCmd.Flags().StringVarP(&bccFilePath, "bcc", "b", bccFilePath, "The path of the BCC source file.")
	createCmd.Flags().StringVarP(&wasmFilePath, "wasm", "w", wasmFilePath, "The path of the WASM binary file.")
}

func parseModuleJsonFile(moduleJsonFilePath string) (*apiserver.CreateModuleReq, error) {
	bytes, err := file.ReadBin(moduleJsonFilePath)
	if err != nil {
		return nil, err
	}
	var moduleReq *apiserver.CreateModuleReq
	err = json.Unmarshal([]byte(bytes), &moduleReq)
	if err != nil {
		return nil, err
	}
	return moduleReq, nil
}
