package randomstore

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"strings"
	"sync"
	"time"

	"github.com/LabGroupware/go-measure-tui/internal/app"
	"gopkg.in/yaml.v3"
)

type RandomStringValueGenerator struct {
	Key     string
	Length  int
	CharSet string
}

var predefinedCharsets = map[string]string{
	"numeric": "0123456789",
	"capital": "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
	"lower":   "abcdefghijklmnopqrstuvwxyz",
}

const defaultCharset = "numeric,capital,lower,[?!]"

const defaultLength = 8

func (p *RandomStringValueGenerator) Generate(ctx context.Context, ctr *app.Container, store *sync.Map) error {
	if p.CharSet == "" {
		p.CharSet = defaultCharset
	}
	charset := buildCharset(p.CharSet)
	if charset == "" {
		return fmt.Errorf("charset is empty")
	}

	if p.Length <= 0 {
		p.Length = defaultLength
	}

	rand.N(time.Now().UnixNano())
	result := make([]byte, p.Length)
	for i := 0; i < p.Length; i++ {
		result[i] = charset[rand.N(len(charset))]
	}

	store.Store(p.Key, string(result))
	return nil
}

func buildCharset(pattern string) string {
	var charsetBuilder strings.Builder
	parts := strings.Split(pattern, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			// Handle custom character sets in square brackets
			charsetBuilder.WriteString(part[1 : len(part)-1])
		} else if val, exists := predefinedCharsets[part]; exists {
			// Use predefined character sets
			charsetBuilder.WriteString(val)
		}
	}
	return charsetBuilder.String()
}

type RandomStoreValueStringDataConfig struct {
	Key     string `yaml:"key"`
	Type    string `yaml:"type"`
	Length  int    `yaml:"length"`
	CharSet string `yaml:"charSet"` // "numeric,[!@#$%^&*],capital,lower"
}

func (p *RandomStoreValueStringDataConfig) Init(conf io.Reader) error {
	decoder := yaml.NewDecoder(conf)
	if err := decoder.Decode(p); err != nil {
		return fmt.Errorf("failed to decode yaml: %w", err)
	}
	return nil
}

func (p *RandomStoreValueStringDataConfig) GeneratorFactory(ctx context.Context, ctr *app.Container) (RadomGenerator, error) {
	return &RandomStringValueGenerator{
		Key:     p.Key,
		Length:  p.Length,
		CharSet: p.CharSet,
	}, nil
}
