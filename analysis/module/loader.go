package module

import (
	"fmt"
	"go-mythril/analysis/module/modules"
	"go-mythril/support"
	"go-mythril/utils"
	"reflect"
)

type ModuleLoader struct {
	Modules []modules.DetectionModule
}

func NewModuleLoader() *ModuleLoader {
	module := make([]modules.DetectionModule, 0)
	loader := &ModuleLoader{
		Modules: module,
	}
	loader.registerMythrilModules()
	return loader
}

func (loader *ModuleLoader) RegisterModule(detectionModule modules.DetectionModule) {
	loader.Modules = append(loader.Modules, detectionModule)
}

func (loader *ModuleLoader) GetDetectionModules(whiteList []string) []modules.DetectionModule {
	result := make([]modules.DetectionModule, len(loader.Modules))
	copy(result, loader.Modules)

	realResult := make([]modules.DetectionModule, 0)
	if len(whiteList) > 0 {
		availableNames := make([]string, 0)
		for _, module := range result {
			availableNames = append(availableNames, reflect.TypeOf(module).String()[9:])
		}

		for _, name := range whiteList {
			if !utils.In(name, availableNames) {
				fmt.Println("Invalid detection module:" + name)
			}
		}

		for _, module := range result {
			if utils.In(reflect.TypeOf(module).String()[9:], whiteList) {
				realResult = append(realResult, module)
			}
		}
	}
	args := support.NewArgs()
	if args.UseIntegerModule == false {
		for index, module := range realResult {
			if reflect.TypeOf(module).String()[9:] == "IntegerArithmetics" {
				realResult = append(realResult[:index], realResult[index+1:]...)
			}
		}
	}
	// TODO: EntryPoints

	return realResult
}

func (loader *ModuleLoader) registerMythrilModules() {
	loader.Modules = append(loader.Modules,
		//modules.NewIntegerArithmetics(),
		modules.NewTxOrigin(),
		//modules.NewPredictableVariables(),
		//modules.NewExternalCalls(),
		//modules.NewStateChangeAfterCall(),
		//modules.NewArbitraryJump(),
		//modules.NewArbitraryStorage(),
		//modules.NewArbitraryDelegateCall(),
		//modules.NewPredictableVariables(),

		//modules.NewEtherThief(),
		//modules.NewExceptions(),
		//modules.NewExternalCalls(),
		//modules.NewIntegerArithmetics(),
		//modules.NewMultipleSends(),
		//modules.NewStateChangeAfterCall(),
		//modules.NewAccidentallyKillable(),
		//modules.NewUncheckedRetval(),
		//modules.NewUserAssertions(),
	)
}
