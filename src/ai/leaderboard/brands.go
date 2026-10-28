package leaderboard

import (
	"regexp"
	"strings"
	"unicode"
)

type brand struct {
	key   string
	name  string
	orgs  []string // normalized organization ids (lowercase, alphanumerics only)
	model *regexp.Regexp
}

// brands maps the organization spellings used by Arena ("anthropic", "zai")
// and Epoch ("Google DeepMind", "Z.ai (Zhipu AI)") to one vendor. The model
// pattern is only a fallback when the source leaves organization empty.
var brands = []brand{
	{"anthropic", "Anthropic", []string{"anthropic"}, regexp.MustCompile(`(?i)^claude`)},
	{"openai", "OpenAI", []string{"openai"}, regexp.MustCompile(`(?i)^(gpt|chatgpt|o[1-9]\b)`)},
	{"google", "Google", []string{"google", "googledeepmind", "deepmind"}, regexp.MustCompile(`(?i)^(gemini|gemma|imagen|veo)`)},
	{"meta", "Meta", []string{"meta", "metaai", "facebook"}, regexp.MustCompile(`(?i)^(llama|muse)`)},
	{"xai", "xAI", []string{"xai"}, regexp.MustCompile(`(?i)^grok`)},
	{"deepseek", "DeepSeek", []string{"deepseek"}, regexp.MustCompile(`(?i)^deepseek`)},
	{"alibaba", "Alibaba", []string{"alibaba", "qwen", "alibabacloud"}, regexp.MustCompile(`(?i)^(qwen|wan)`)},
	{"moonshot", "Moonshot AI", []string{"moonshot", "moonshotai", "kimi"}, regexp.MustCompile(`(?i)^kimi`)},
	{"zai", "Z.ai", []string{"zai", "zaizhipuai", "zhipu", "zhipuai"}, regexp.MustCompile(`(?i)^(glm|chatglm)`)},
	{"minimax", "MiniMax", []string{"minimax"}, regexp.MustCompile(`(?i)^minimax`)},
	{"mistral", "Mistral", []string{"mistral", "mistralai"}, regexp.MustCompile(`(?i)^(mistral|codestral|magistral)`)},
	{"xiaomi", "Xiaomi", []string{"xiaomi"}, regexp.MustCompile(`(?i)^mimo`)},
	{"microsoft", "Microsoft", []string{"microsoft", "microsoftai", "microsoftresearch"}, regexp.MustCompile(`(?i)^(mai|phi)`)},
	{"bytedance", "ByteDance", []string{"bytedance", "bytedanceseed"}, regexp.MustCompile(`(?i)^(seed|dreamina|doubao)`)},
	{"tencent", "Tencent", []string{"tencent"}, regexp.MustCompile(`(?i)^hunyuan`)},
	{"nvidia", "NVIDIA", []string{"nvidia"}, regexp.MustCompile(`(?i)^nemotron`)},
	{"bfl", "Black Forest Labs", []string{"bfl", "blackforestlabs"}, regexp.MustCompile(`(?i)^flux`)},
	{"stepfun", "StepFun", []string{"stepfun"}, nil},
	{"baidu", "Baidu", []string{"baidu"}, regexp.MustCompile(`(?i)^ernie`)},
	{"luma", "Luma AI", []string{"lumaai", "luma"}, nil},
}

var variantPattern = regexp.MustCompile(`^(.*\S)\s+\(([^()]+)\)$`)

// splitVariant turns "Claude Opus 5.5 (High)" into ("Claude Opus 5.5", "High").
func splitVariant(name string) (string, string) {
	if m := variantPattern.FindStringSubmatch(strings.TrimSpace(name)); m != nil {
		return m[1], m[2]
	}
	return strings.TrimSpace(name), ""
}

// brandOf returns a vendor key and a display organization name.
func brandOf(org, model string) (string, string) {
	org = strings.TrimSpace(org)
	if i := strings.Index(org, ","); i > 0 { // Epoch lists co-developers: "Google DeepMind,Google"
		org = strings.TrimSpace(org[:i])
	}
	id := normalize(org)
	for _, b := range brands {
		for _, o := range b.orgs {
			if id == o {
				return b.key, b.name
			}
		}
	}
	if org == "" {
		for _, b := range brands {
			if b.model != nil && b.model.MatchString(model) {
				return b.key, b.name
			}
		}
		return "", ""
	}
	return id, prettyOrg(org)
}

func normalize(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// prettyOrg title-cases Arena's lowercase ids ("kling" → "Kling") and leaves
// properly cased names from other sources alone.
func prettyOrg(org string) string {
	if org != strings.ToLower(org) {
		return org
	}
	words := strings.FieldsFunc(org, func(r rune) bool { return r == '-' || r == '_' || r == ' ' })
	for i, w := range words {
		r := []rune(w)
		words[i] = string(unicode.ToUpper(r[0])) + string(r[1:])
	}
	return strings.Join(words, " ")
}
