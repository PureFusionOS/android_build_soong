// Copyright 2016 Google Inc. All rights reserved.
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

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"android/soong/android"
)

var (
	// Flags used by lots of devices.  Putting them in package static variables
	// will save bytes in build.ninja so they aren't repeated for every file
	commonGlobalCflags = []string{
		"-DANDROID",
		"-fmessage-length=0",
		"-W",
		"-Wall",
		"-Wno-unused",
		"-Winit-self",
		"-Wpointer-arith",
		"-Wno-address-of-packed-member",
		"-Wno-main",
		"-Wno-instantiation-after-specialization",
		"-Wno-max-unsigned-zero",

		// Make paths in deps files relative
		"-no-canonical-prefixes",

		"-DNDEBUG",
		"-UDEBUG",

		"-fno-exceptions",
		"-Wno-multichar",

		"-O2",
		"-g",

		"-fno-strict-aliasing",
	}

	commonGlobalConlyflags = []string{}

	deviceGlobalCflags = []string{
		"-fdiagnostics-color",

		"-fno-canonical-system-headers",
		"-ffunction-sections",
		"-funwind-tables",
		"-fstack-protector-strong",
		"-Wa,--noexecstack",
		"-D_FORTIFY_SOURCE=2",

		"-Wstrict-aliasing=2",

		"-Wno-error=return-type",
		"-Wno-error=non-virtual-dtor",
		"-Wno-error=address",
		"-Wno-error=sequence-point",
		"-Wno-error=date-time",
		"-Werror=format-security",
	}

	deviceGlobalLdflags = []string{
		"-Wl,-z,noexecstack",
		"-Wl,-z,relro",
		"-Wl,-z,now",
		"-Wl,--build-id=md5",
		"-Wl,--warn-shared-textrel",
		"-Wl,--fatal-warnings",
		"-Wl,--no-undefined-version",
	}

	// Maybe in the future, LLD will have different global flags.
	deviceGlobalLldflags = []string{
		"-Wl,-z,noexecstack",
		"-Wl,-z,relro",
		"-Wl,-z,now",
		"-Wl,--build-id=md5",
		"-Wl,--warn-shared-textrel",
		"-Wl,--fatal-warnings",
		"-Wl,--no-undefined-version",
	}

	hostGlobalCflags = []string{}

	hostGlobalLdflags = []string{}

	hostGlobalLldflags = []string{}

	commonGlobalCppflags = []string{
		"-Wno-inconsistent-missing-override",
		"-Wsign-promo",
	}

	noOverrideGlobalCflags = []string{
		"-Wno-error=int-to-pointer-cast",
		"-Wno-error=pointer-to-int-cast",
	}

	IllegalFlags = []string{
		"-w",
	}

	CStdVersion               = "gnu99"
	CppStdVersion             = "gnu++14"
	GccCppStdVersion          = "gnu++11"
	ExperimentalCStdVersion   = "gnu11"
	ExperimentalCppStdVersion = "gnu++1z"
	SDClang                   = true

	NdkMaxPrebuiltVersionInt = 24

	// prebuilts/clang default settings.
	ClangDefaultBase         = "prebuilts/clang/host"
	ClangDefaultVersion      = "6.0"
	ClangDefaultShortVersion = "6.0"
)

var pctx = android.NewPackageContext("android/soong/cc/config")

func init() {
	if android.BuildOs == android.Linux {
		commonGlobalCflags = append(commonGlobalCflags, "-fdebug-prefix-map=/proc/self/cwd=")
	}

	pctx.StaticVariable("CommonGlobalCflags", strings.Join(commonGlobalCflags, " "))
	pctx.StaticVariable("CommonGlobalConlyflags", strings.Join(commonGlobalConlyflags, " "))
	pctx.StaticVariable("DeviceGlobalCflags", strings.Join(deviceGlobalCflags, " "))
	pctx.StaticVariable("DeviceGlobalLdflags", strings.Join(deviceGlobalLdflags, " "))
	pctx.StaticVariable("DeviceGlobalLldflags", strings.Join(deviceGlobalLldflags, " "))
	pctx.StaticVariable("HostGlobalCflags", strings.Join(hostGlobalCflags, " "))
	pctx.StaticVariable("HostGlobalLdflags", strings.Join(hostGlobalLdflags, " "))
	pctx.StaticVariable("HostGlobalLldflags", strings.Join(hostGlobalLldflags, " "))
	pctx.StaticVariable("NoOverrideGlobalCflags", strings.Join(noOverrideGlobalCflags, " "))

	pctx.StaticVariable("CommonGlobalCppflags", strings.Join(commonGlobalCppflags, " "))

	pctx.StaticVariable("CommonClangGlobalCflags",
		strings.Join(append(ClangFilterUnknownCflags(commonGlobalCflags), "${ClangExtraCflags}"), " "))
	pctx.StaticVariable("DeviceClangGlobalCflags",
		strings.Join(append(ClangFilterUnknownCflags(deviceGlobalCflags), "${ClangExtraTargetCflags}"), " "))
	pctx.StaticVariable("HostClangGlobalCflags",
		strings.Join(ClangFilterUnknownCflags(hostGlobalCflags), " "))
	pctx.StaticVariable("NoOverrideClangGlobalCflags",
		strings.Join(append(ClangFilterUnknownCflags(noOverrideGlobalCflags), "${ClangExtraNoOverrideCflags}"), " "))

	pctx.StaticVariable("CommonClangGlobalCppflags",
		strings.Join(append(ClangFilterUnknownCflags(commonGlobalCppflags), "${ClangExtraCppflags}"), " "))

	// Everything in these lists is a crime against abstraction and dependency tracking.
	// Do not add anything to this list.
	pctx.PrefixedExistentPathsForSourcesVariable("CommonGlobalIncludes", "-I",
		[]string{
			"system/core/include",
			"system/media/audio/include",
			"hardware/libhardware/include",
			"hardware/libhardware_legacy/include",
			"hardware/ril/include",
			"libnativehelper/include",
			"frameworks/native/include",
			"frameworks/native/opengl/include",
			"frameworks/av/include",
		})
	// This is used by non-NDK modules to get jni.h. export_include_dirs doesn't help
	// with this, since there is no associated library.
	pctx.PrefixedExistentPathsForSourcesVariable("CommonNativehelperInclude", "-I",
		[]string{"libnativehelper/include_deprecated"})

	pctx.SourcePathVariable("ClangDefaultBase", ClangDefaultBase)
	pctx.VariableFunc("ClangBase", func(config interface{}) (string, error) {
		if override := config.(android.Config).Getenv("6.0"); override != "" {
			return override, nil
		}
		return "${ClangDefaultBase}", nil
	})
	pctx.VariableFunc("ClangVersion", func(config interface{}) (string, error) {
		if override := config.(android.Config).Getenv("6.0"); override != "" {
			return override, nil
		}
		return "6.0", nil
	})
	pctx.StaticVariable("ClangPath", "${ClangBase}/${HostPrebuiltTag}/${ClangVersion}")
	pctx.StaticVariable("ClangBin", "${ClangPath}/bin")

	pctx.VariableFunc("ClangShortVersion", func(config interface{}) (string, error) {
		if override := config.(android.Config).Getenv("6.0"); override != "" {
			return override, nil
		}
		return "6.0", nil
	})
	pctx.StaticVariable("ClangAsanLibDir", "${ClangPath}/lib/clang/6.0.0/lib/linux")
	pctx.StaticVariable("LLVMGoldPlugin", "${ClangPath}/lib/LLVMgold.so")

	// These are tied to the version of LLVM directly in external/llvm, so they might trail the host prebuilts
	// being used for the rest of the build process.
	pctx.SourcePathVariable("RSClangBase", "prebuilts/clang/host")
	pctx.SourcePathVariable("RSClangVersion", "6.0")
	pctx.SourcePathVariable("RSReleaseVersion", "6.0")
	pctx.StaticVariable("RSLLVMPrebuiltsPath", "${RSClangBase}/${HostPrebuiltTag}/${RSClangVersion}/bin")
	pctx.StaticVariable("RSIncludePath", "${RSClangBase}/${HostPrebuiltTag}/${RSClangVersion}/lib/clang/6.0.0/include")

	pctx.PrefixedExistentPathsForSourcesVariable("RsGlobalIncludes", "-I",
		[]string{
			"external/clang/lib/Headers",
			"frameworks/rs/script_api/include",
		})

	pctx.VariableFunc("CcWrapper", func(config interface{}) (string, error) {
		if override := config.(android.Config).Getenv("CC_WRAPPER"); override != "" {
			return override + " ", nil
		}
		return "", nil
	})

	setSdclangVars()
}

func setSdclangVars() {
	sdclangPath := os.Getenv("SDCLANG_PATH")
	sdclangPath2 := os.Getenv("SDCLANG_PATH2")
	sdclangAEFlag := ""
	sdclangFlags := ""
	sdclangFlags2 := ""

	product := os.Getenv("TARGET_PRODUCT")
	androidRoot := os.Getenv("ANDROID_BUILD_TOP")
	aeConfigPath := os.Getenv("SDCLANG_AE_CONFIG")
	sdclangConfigPath := os.Getenv("SDCLANG_CONFIG")

	type sdclangAEConfig struct {
		SDCLANG_AE_FLAG string
	}

	// Load AE config file and set AE flag
	aeConfigFile := path.Join(androidRoot, aeConfigPath)
	if file, err := os.Open(aeConfigFile); err == nil {
		decoder := json.NewDecoder(file)
		aeConfig := sdclangAEConfig{}
		if err := decoder.Decode(&aeConfig); err == nil {
			sdclangAEFlag = aeConfig.SDCLANG_AE_FLAG
		} else {
//			panic(err)
		}
	}

	// Load SD Clang config file and set SD Clang variables
	sdclangConfigFile := path.Join(androidRoot, sdclangConfigPath)
	var sdclangConfig interface{}
	if file, err := os.Open(sdclangConfigFile); err == nil {
		decoder := json.NewDecoder(file)
                // Parse the config file
		if err := decoder.Decode(&sdclangConfig); err == nil {
			config := sdclangConfig.(map[string]interface{})
			// Retrieve the default block
			if dev, ok := config["default"]; ok {
				devConfig := dev.(map[string]interface{})
				// SDCLANG is optional in the default block
				if _, ok := devConfig["SDCLANG"]; ok {
					SDClang = devConfig["SDCLANG"].(bool)
				}
				// SDCLANG_PATH is required in the default block
				if _, ok := devConfig["SDCLANG_PATH"]; ok {
					sdclangPath = devConfig["SDCLANG_PATH"].(string)
				} else {
					panic("SDCLANG_PATH is required in the default block")
				}
				// SDCLANG_PATH_2 is required in the default block
				if _, ok := devConfig["SDCLANG_PATH_2"]; ok {
					sdclangPath2 = devConfig["SDCLANG_PATH_2"].(string)
				} else {
					panic("SDCLANG_PATH_2 is required in the default block")
				}
				// SDCLANG_FLAGS is optional in the default block
				if _, ok := devConfig["SDCLANG_FLAGS"]; ok {
					sdclangFlags = devConfig["SDCLANG_FLAGS"].(string)
				}
				// SDCLANG_FLAGS_2 is optional in the default block
				if _, ok := devConfig["SDCLANG_FLAGS_2"]; ok {
					sdclangFlags2 = devConfig["SDCLANG_FLAGS_2"].(string)
				}
			} else {
				panic("Default block is required in the SD Clang config file")
			}
			// Retrieve the device specific block if it exists in the config file
			if dev, ok := config[product]; ok {
				devConfig := dev.(map[string]interface{})
				// SDCLANG is optional in the device specific block
				if _, ok := devConfig["SDCLANG"]; ok {
					SDClang = devConfig["SDCLANG"].(bool)
				}
				// SDCLANG_PATH is optional in the device specific block
				if _, ok := devConfig["SDCLANG_PATH"]; ok {
					sdclangPath = devConfig["SDCLANG_PATH"].(string)
				}
				// SDCLANG_PATH_2 is optional in the device specific block
				if _, ok := devConfig["SDCLANG_PATH_2"]; ok {
					sdclangPath2 = devConfig["SDCLANG_PATH_2"].(string)
				}
				// SDCLANG_FLAGS is optional in the device specific block
				if _, ok := devConfig["SDCLANG_FLAGS"]; ok {
					sdclangFlags = devConfig["SDCLANG_FLAGS"].(string)
				}
				// SDCLANG_FLAGS_2 is optional in the device specific block
				if _, ok := devConfig["SDCLANG_FLAGS_2"]; ok {
					sdclangFlags2 = devConfig["SDCLANG_FLAGS_2"].(string)
				}
			}
		} else {
//			panic(err)
		}
	} else {
//		fmt.Println(err)
	}

	// Override SDCLANG if the varialbe is set in the environment
	if sdclang := os.Getenv("SDCLANG"); sdclang != "" {
		if override, err := strconv.ParseBool(sdclang); err == nil {
			SDClang = override
		}
	}

	// Sanity check SDCLANG_PATH
	if envPath := os.Getenv("SDCLANG_PATH"); sdclangPath == "" && envPath == "" {
		panic("SDCLANG_PATH can not be empty")
	}

	// Sanity check SDCLANG_PATH_2
	if envPath := os.Getenv("SDCLANG_PATH_2"); sdclangPath2 == "" && envPath == "" {
		panic("SDCLANG_PATH_2 can not be empty")
	}

	// Override SDCLANG_PATH if the variable is set in the environment
	pctx.VariableFunc("SDClangBin", func(config interface{}) (string, error) {
		if override := config.(android.Config).Getenv("SDCLANG_PATH"); override != "" {
			return override, nil
		}
		return sdclangPath, nil
	})

	// Override SDCLANG_PATH_2 if the variable is set in the environment
	pctx.VariableFunc("SDClangBin2", func(config interface{}) (string, error) {
		if override := config.(android.Config).Getenv("SDCLANG_PATH_2"); override != "" {
			return override, nil
		}
		return sdclangPath2, nil
	})

	// Override SDCLANG_COMMON_FLAGS if the variable is set in the environment
	pctx.VariableFunc("SDClangFlags", func(config interface{}) (string, error) {
		if override := config.(android.Config).Getenv("SDCLANG_COMMON_FLAGS"); override != "" {
			return override, nil
		}
		return sdclangAEFlag + " " + sdclangFlags, nil
	})

	// Override SDCLANG_COMMON_FLAGS_2 if the variable is set in the environment
	pctx.VariableFunc("SDClangFlags2", func(config interface{}) (string, error) {
		if override := config.(android.Config).Getenv("SDCLANG_COMMON_FLAGS_2"); override != "" {
			return override, nil
		}
		return sdclangAEFlag + " " + sdclangFlags2, nil
	})
}

var HostPrebuiltTag = pctx.VariableConfigMethod("HostPrebuiltTag", android.Config.PrebuiltOS)

func bionicHeaders(bionicArch, kernelArch string) string {
	return strings.Join([]string{
		"-isystem bionic/libc/arch-" + bionicArch + "/include",
		"-isystem bionic/libc/include",
		"-isystem bionic/libc/kernel/uapi",
		"-isystem bionic/libc/kernel/uapi/asm-" + kernelArch,
		"-isystem bionic/libc/kernel/android/scsi",
		"-isystem bionic/libc/kernel/android/uapi",
	}, " ")
}

func replaceFirst(slice []string, from, to string) {
	if slice[0] != from {
		panic(fmt.Errorf("Expected %q, found %q", from, to))
	}
	slice[0] = to
}
