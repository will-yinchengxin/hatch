package leaderboard

// Only sources that publish current, CC BY 4.0 data anonymously are listed.
// Arena boards use the style-controlled subsets because that is arena.ai's
// default view since 2025-05; raw subsets would not match the official page.
const (
	arenaBase = "https://huggingface.co/datasets/lmarena-ai/leaderboard-dataset/resolve/main"
	epochURL  = "https://epoch.ai/data/benchmark_data.zip"
)

// Score formats tell the UI how to print and scale a board's numbers.
const (
	FormatElo     = "elo"     // Bradley–Terry rating, interval scale (zero is meaningless)
	FormatIPS     = "ips"     // Agent Arena IPS score
	FormatIndex   = "index"   // Epoch Capabilities Index
	FormatPercent = "percent" // accuracy 0–100
)

type feed struct {
	id  string
	url string
}

var feeds = []feed{
	{"arena/text_style_control", arenaBase + "/text_style_control/latest-00000-of-00001.parquet"},
	{"arena/agent", arenaBase + "/agent/latest-00000-of-00001.parquet"},
	{"arena/webdev", arenaBase + "/webdev/latest-00000-of-00001.parquet"},
	{"arena/text_to_image", arenaBase + "/text_to_image/latest-00000-of-00001.parquet"},
	{"arena/image_edit", arenaBase + "/image_edit/latest-00000-of-00001.parquet"},
	{"arena/vision_style_control", arenaBase + "/vision_style_control/latest-00000-of-00001.parquet"},
	{"epoch", epochURL},
}

// Definition is the static description of one board; pick is the Arena
// category or the Epoch CSV inside the archive.
type Definition struct {
	ID      string `json:"id"`
	Domain  string `json:"domain"`
	Title   string `json:"title"`
	Tag     string `json:"tag"`
	Summary string `json:"summary"`
	Source  string `json:"source"`
	URL     string `json:"url"`
	License string `json:"license"`
	Format  string `json:"format"`
	Unit    string `json:"unit"`
	feed    string
	pick    string
}

func arena(id, domain, title, tag, summary, url, feedID, category, format string) Definition {
	unit := "Arena 分"
	if format == FormatIPS {
		unit = "IPS"
	}
	return Definition{ID: id, Domain: domain, Title: title, Tag: tag, Summary: summary, Source: "Arena", URL: url,
		License: "CC BY 4.0", Format: format, Unit: unit, feed: feedID, pick: category}
}

func epoch(id, domain, title, tag, summary, url, csv, format, unit string) Definition {
	return Definition{ID: id, Domain: domain, Title: title, Tag: tag, Summary: summary, Source: "Epoch AI", URL: url,
		License: "CC BY 4.0", Format: format, Unit: unit, feed: "epoch", pick: csv}
}

// Definitions are ordered by domain, then by how representative the board is.
var definitions = []Definition{
	arena("arena-coding", "coding", "Text Arena · 编程", "人类偏好",
		"真人匿名盲测中编程类对话的胜率评分，已启用风格控制，与 arena.ai 默认榜一致。",
		"https://arena.ai/leaderboard/chat/text/coding", "arena/text_style_control", "coding", FormatElo),
	arena("arena-agent", "coding", "Agent Arena", "智能体编程",
		"真实编码智能体会话中的模型表现（IPS 评分），综合任务完成、工具调用与可控性信号。",
		"https://arena.ai/leaderboard", "arena/agent", "overall", FormatIPS),
	arena("arena-webdev", "frontend", "Code Arena · WebDev", "网页生成",
		"真人对比两个模型按同一需求生成的可交互网页，投票得出评分。",
		"https://arena.ai/leaderboard/code/webdev", "arena/webdev", "overall", FormatElo),
	arena("arena-image-to-webdev", "frontend", "Code Arena · 截图转网页", "设计还原",
		"根据截图或设计稿生成网页，真人投票比较还原度与可用性。",
		"https://arena.ai/leaderboard/code/webdev", "arena/webdev", "image_to_webdev", FormatElo),
	arena("arena-text-to-image", "image", "Text-to-Image Arena", "文生图",
		"同一提示词下两张生成图匿名对比，真人投票得出评分。",
		"https://arena.ai/leaderboard/image/text-to-image", "arena/text_to_image", "overall", FormatElo),
	arena("arena-image-edit", "image", "Image Edit Arena", "图像编辑",
		"按指令修改图片，真人比较编辑结果的准确度与质量。",
		"https://arena.ai/leaderboard/image/image-edit", "arena/image_edit", "overall", FormatElo),
	arena("arena-vision", "image", "Vision Arena", "视觉理解",
		"带图片输入的对话中比较模型的理解与回答，已启用风格控制。",
		"https://arena.ai/leaderboard/chat/vision", "arena/vision_style_control", "overall", FormatElo),
	arena("arena-text", "general", "Text Arena · 总榜", "人类偏好",
		"真人匿名盲测投票得出的综合对话评分，已启用风格控制，与 arena.ai 默认榜一致。",
		"https://arena.ai/leaderboard/chat/text", "arena/text_style_control", "overall", FormatElo),
	arena("arena-chinese", "general", "Text Arena · 中文", "中文",
		"仅统计中文提问的盲测投票，已启用风格控制。",
		"https://arena.ai/leaderboard/chat/text", "arena/text_style_control", "chinese", FormatElo),
	epoch("epoch-eci", "general", "Epoch Capabilities Index", "综合能力指数",
		"Epoch AI 将数十项基准拟合到同一能力刻度上的综合指数，附置信区间。",
		"https://epoch.ai/benchmarks/eci", "eci", FormatIndex, "ECI"),
	epoch("epoch-gpqa", "general", "GPQA Diamond", "科学推理",
		"Epoch AI 自行运行的研究生级物理、化学、生物选择题，展示平均准确率。",
		"https://epoch.ai/benchmarks/gpqa-diamond", "gpqa_diamond.csv", FormatPercent, "准确率"),
}
