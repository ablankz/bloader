package randomstore

import (
	"context"
	"io"

	"github.com/LabGroupware/go-measure-tui/internal/app"
)

type RandomGenerator interface {
	Init(conf io.Reader) error
	GeneratorFactory(ctx context.Context, ctr *app.Container) (RadomGenerator, error)
}

func GetRandomGeneratorFactory(t string) RandomGenerator {
	switch t {
	case "constant":
		return &RandomStoreValueConstantDataConfig{}
	case "element":
		return &RandomStoreValueElementDataConfig{}
	case "int":
		return &RandomStoreValueIntDataConfig{}
	case "float":
		return &RandomStoreValueFloatDataConfig{}
	case "string":
		return &RandomStoreValueStringDataConfig{}
	case "bool":
		return &RandomStoreValueBoolDataConfig{}
	case "uuid":
		return &RandomStoreValueUUIDDataConfig{}
	case "datetime":
		return &RandomStoreValueDatetimeDataConfig{}
	default:
		return nil
	}
}
