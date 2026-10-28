export namespace agent {
	
	export class ClaudeCodeConfig {
	    model: string;
	    effortLevel: string;
	    effectiveModel: string;
	
	    static createFrom(source: any = {}) {
	        return new ClaudeCodeConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = source["model"];
	        this.effortLevel = source["effortLevel"];
	        this.effectiveModel = source["effectiveModel"];
	    }
	}
	export class ClaudeCodeStatus {
	    claudeInstalled: boolean;
	    claudeVersion: string;
	    vpnConnected: boolean;
	    vpnInterfaces: string[];
	    hasUsageData: boolean;
	    fiveHourPct: number;
	    fiveHourReset: string;
	    sevenDayPct: number;
	    sevenDayReset: string;
	    planType: string;
	    usageUpdatedAt: string;
	    usageAge: string;
	    usageStale: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ClaudeCodeStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.claudeInstalled = source["claudeInstalled"];
	        this.claudeVersion = source["claudeVersion"];
	        this.vpnConnected = source["vpnConnected"];
	        this.vpnInterfaces = source["vpnInterfaces"];
	        this.hasUsageData = source["hasUsageData"];
	        this.fiveHourPct = source["fiveHourPct"];
	        this.fiveHourReset = source["fiveHourReset"];
	        this.sevenDayPct = source["sevenDayPct"];
	        this.sevenDayReset = source["sevenDayReset"];
	        this.planType = source["planType"];
	        this.usageUpdatedAt = source["usageUpdatedAt"];
	        this.usageAge = source["usageAge"];
	        this.usageStale = source["usageStale"];
	    }
	}
	export class CodexModelConfig {
	    model: string;
	    effortLevel: string;
	    goalsEnabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CodexModelConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = source["model"];
	        this.effortLevel = source["effortLevel"];
	        this.goalsEnabled = source["goalsEnabled"];
	    }
	}
	export class CodexUsageWindow {
	    label: string;
	    usedPercent: number;
	    leftPercent: number;
	    windowMinutes: number;
	    resetsAt: number;
	    resetIn: string;
	    available: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CodexUsageWindow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.usedPercent = source["usedPercent"];
	        this.leftPercent = source["leftPercent"];
	        this.windowMinutes = source["windowMinutes"];
	        this.resetsAt = source["resetsAt"];
	        this.resetIn = source["resetIn"];
	        this.available = source["available"];
	    }
	}
	export class CodexUsageStatus {
	    codexInstalled: boolean;
	    codexVersion: string;
	    vpnConnected: boolean;
	    vpnInterfaces: string[];
	    hasUsageData: boolean;
	    planType: string;
	    limitReached: string;
	    snapshotAt: string;
	    primary: CodexUsageWindow;
	    secondary: CodexUsageWindow;
	
	    static createFrom(source: any = {}) {
	        return new CodexUsageStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.codexInstalled = source["codexInstalled"];
	        this.codexVersion = source["codexVersion"];
	        this.vpnConnected = source["vpnConnected"];
	        this.vpnInterfaces = source["vpnInterfaces"];
	        this.hasUsageData = source["hasUsageData"];
	        this.planType = source["planType"];
	        this.limitReached = source["limitReached"];
	        this.snapshotAt = source["snapshotAt"];
	        this.primary = this.convertValues(source["primary"], CodexUsageWindow);
	        this.secondary = this.convertValues(source["secondary"], CodexUsageWindow);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace app {
	
	export class AIProvidersResult {
	    providers: config.AIProvider[];
	    activeProviderId: string;
	
	    static createFrom(source: any = {}) {
	        return new AIProvidersResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.providers = this.convertValues(source["providers"], config.AIProvider);
	        this.activeProviderId = source["activeProviderId"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AIScriptSuggestion {
	    name: string;
	    desc: string;
	    type: string;
	    commands: string[];
	    pidFile: string;
	    needsArgs: boolean;
	    args: config.ArgDef[];
	    tags: string[];
	    presets: config.CommandPreset[];
	
	    static createFrom(source: any = {}) {
	        return new AIScriptSuggestion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.desc = source["desc"];
	        this.type = source["type"];
	        this.commands = source["commands"];
	        this.pidFile = source["pidFile"];
	        this.needsArgs = source["needsArgs"];
	        this.args = this.convertValues(source["args"], config.ArgDef);
	        this.tags = source["tags"];
	        this.presets = this.convertValues(source["presets"], config.CommandPreset);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AppDataPaths {
	    traceLogDir: string;
	    traceLogDirIsSet: boolean;
	    skillsDir: string;
	    skillsDirIsSet: boolean;
	    defaultTraceDir: string;
	    defaultSkillsDir: string;
	
	    static createFrom(source: any = {}) {
	        return new AppDataPaths(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.traceLogDir = source["traceLogDir"];
	        this.traceLogDirIsSet = source["traceLogDirIsSet"];
	        this.skillsDir = source["skillsDir"];
	        this.skillsDirIsSet = source["skillsDirIsSet"];
	        this.defaultTraceDir = source["defaultTraceDir"];
	        this.defaultSkillsDir = source["defaultSkillsDir"];
	    }
	}
	export class RetentionView {
	    traceRetainDays: number;
	    runsRetainDays: number;
	    runsKeepPerName: number;
	
	    static createFrom(source: any = {}) {
	        return new RetentionView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.traceRetainDays = source["traceRetainDays"];
	        this.runsRetainDays = source["runsRetainDays"];
	        this.runsKeepPerName = source["runsKeepPerName"];
	    }
	}
	export class ScheduleInfo {
	    path: string;
	    schedule: string;
	    nextRun: number;
	
	    static createFrom(source: any = {}) {
	        return new ScheduleInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.schedule = source["schedule"];
	        this.nextRun = source["nextRun"];
	    }
	}
	export class ZshrcApplyResult {
	    zshrcPath: string;
	    backupPath: string;
	
	    static createFrom(source: any = {}) {
	        return new ZshrcApplyResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.zshrcPath = source["zshrcPath"];
	        this.backupPath = source["backupPath"];
	    }
	}
	export class ZshrcInfo {
	    shell: string;
	    isZsh: boolean;
	    zshrcPath: string;
	    hatchLines: string[];
	
	    static createFrom(source: any = {}) {
	        return new ZshrcInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.shell = source["shell"];
	        this.isZsh = source["isZsh"];
	        this.zshrcPath = source["zshrcPath"];
	        this.hatchLines = source["hatchLines"];
	    }
	}
	export class ZshrcOrganizationSuggestion {
	    summary: string[];
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new ZshrcOrganizationSuggestion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.summary = source["summary"];
	        this.content = source["content"];
	    }
	}

}

export namespace backup {
	
	export class Result {
	    path: string;
	    size: number;
	    numFiles: number;
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.size = source["size"];
	        this.numFiles = source["numFiles"];
	    }
	}

}

export namespace config {
	
	export class AIConfig {
	    enabled: boolean;
	    baseUrl: string;
	    apiKey: string;
	    model: string;
	    traceRetainDays?: number;
	
	    static createFrom(source: any = {}) {
	        return new AIConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.baseUrl = source["baseUrl"];
	        this.apiKey = source["apiKey"];
	        this.model = source["model"];
	        this.traceRetainDays = source["traceRetainDays"];
	    }
	}
	export class AIProvider {
	    id: string;
	    name: string;
	    baseUrl: string;
	    apiKey: string;
	    model: string;
	
	    static createFrom(source: any = {}) {
	        return new AIProvider(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.baseUrl = source["baseUrl"];
	        this.apiKey = source["apiKey"];
	        this.model = source["model"];
	    }
	}
	export class AgentResourceSelection {
	    id: string;
	    claude: boolean;
	    codex: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AgentResourceSelection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.claude = source["claude"];
	        this.codex = source["codex"];
	    }
	}
	export class AIWorkspaceProfile {
	    id: string;
	    name: string;
	    rootPath: string;
	    claudeModel?: string;
	    claudeEffort?: string;
	    codexModel?: string;
	    codexEffort?: string;
	    codexGoals: boolean;
	    skills?: AgentResourceSelection[];
	    mcpServers?: AgentResourceSelection[];
	    hooks?: AgentResourceSelection[];
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new AIWorkspaceProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.rootPath = source["rootPath"];
	        this.claudeModel = source["claudeModel"];
	        this.claudeEffort = source["claudeEffort"];
	        this.codexModel = source["codexModel"];
	        this.codexEffort = source["codexEffort"];
	        this.codexGoals = source["codexGoals"];
	        this.skills = this.convertValues(source["skills"], AgentResourceSelection);
	        this.mcpServers = this.convertValues(source["mcpServers"], AgentResourceSelection);
	        this.hooks = this.convertValues(source["hooks"], AgentResourceSelection);
	        this.updatedAt = source["updatedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class ArgDef {
	    name: string;
	    flag: string;
	    default: string;
	    required: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ArgDef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.flag = source["flag"];
	        this.default = source["default"];
	        this.required = source["required"];
	    }
	}
	export class ClaudeCodePreset {
	    id: string;
	    name: string;
	    model: string;
	    effortLevel: string;
	
	    static createFrom(source: any = {}) {
	        return new ClaudeCodePreset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.model = source["model"];
	        this.effortLevel = source["effortLevel"];
	    }
	}
	export class CodexPreset {
	    id: string;
	    name: string;
	    model: string;
	    effortLevel: string;
	
	    static createFrom(source: any = {}) {
	        return new CodexPreset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.model = source["model"];
	        this.effortLevel = source["effortLevel"];
	    }
	}
	export class CommandPreset {
	    name: string;
	    args: string;
	
	    static createFrom(source: any = {}) {
	        return new CommandPreset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.args = source["args"];
	    }
	}
	export class HookDef {
	    id: string;
	    name: string;
	    description?: string;
	    event: string;
	    matcher?: string;
	    command: string;
	    args?: string[];
	    timeout?: number;
	    statusMessage?: string;
	    enabledInClaude?: boolean;
	    enabledInCodex?: boolean;
	    updatedAt?: number;
	
	    static createFrom(source: any = {}) {
	        return new HookDef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.event = source["event"];
	        this.matcher = source["matcher"];
	        this.command = source["command"];
	        this.args = source["args"];
	        this.timeout = source["timeout"];
	        this.statusMessage = source["statusMessage"];
	        this.enabledInClaude = source["enabledInClaude"];
	        this.enabledInCodex = source["enabledInCodex"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class MCPTool {
	    name: string;
	    description?: string;
	
	    static createFrom(source: any = {}) {
	        return new MCPTool(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	    }
	}
	export class MCPProbe {
	    status: string;
	    summary: string;
	    serverName?: string;
	    serverVersion?: string;
	    protocolVersion?: string;
	    tools?: MCPTool[];
	    latencyMs?: number;
	    stderrTail?: string;
	    checkedAt?: number;
	
	    static createFrom(source: any = {}) {
	        return new MCPProbe(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.summary = source["summary"];
	        this.serverName = source["serverName"];
	        this.serverVersion = source["serverVersion"];
	        this.protocolVersion = source["protocolVersion"];
	        this.tools = this.convertValues(source["tools"], MCPTool);
	        this.latencyMs = source["latencyMs"];
	        this.stderrTail = source["stderrTail"];
	        this.checkedAt = source["checkedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MCPMeta {
	    name: string;
	    origin?: string;
	    notes?: string;
	    parked?: Record<string, string>;
	    lastProbe?: Record<string, MCPProbe>;
	
	    static createFrom(source: any = {}) {
	        return new MCPMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.origin = source["origin"];
	        this.notes = source["notes"];
	        this.parked = source["parked"];
	        this.lastProbe = this.convertValues(source["lastProbe"], MCPProbe, true);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LegacyMCPServer {
	    id: string;
	    name: string;
	    description?: string;
	    transport: string;
	    command?: string;
	    args?: string[];
	    cwd?: string;
	    url?: string;
	    env?: Record<string, string>;
	    httpHeaders?: Record<string, string>;
	    bearerTokenEnvVar?: string;
	    startupTimeoutSec?: number;
	    toolTimeoutSec?: number;
	    enabledInClaude?: boolean;
	    enabledInCodex?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LegacyMCPServer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.transport = source["transport"];
	        this.command = source["command"];
	        this.args = source["args"];
	        this.cwd = source["cwd"];
	        this.url = source["url"];
	        this.env = source["env"];
	        this.httpHeaders = source["httpHeaders"];
	        this.bearerTokenEnvVar = source["bearerTokenEnvVar"];
	        this.startupTimeoutSec = source["startupTimeoutSec"];
	        this.toolTimeoutSec = source["toolTimeoutSec"];
	        this.enabledInClaude = source["enabledInClaude"];
	        this.enabledInCodex = source["enabledInCodex"];
	    }
	}
	export class MemoSettings {
	    filesDir?: string;
	    retainDays?: number;
	    autoCleanup?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MemoSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filesDir = source["filesDir"];
	        this.retainDays = source["retainDays"];
	        this.autoCleanup = source["autoCleanup"];
	    }
	}
	export class DataRetention {
	    runsRetainDays?: number;
	    runsKeepPerName?: number;
	
	    static createFrom(source: any = {}) {
	        return new DataRetention(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.runsRetainDays = source["runsRetainDays"];
	        this.runsKeepPerName = source["runsKeepPerName"];
	    }
	}
	export class DocLink {
	    path: string;
	    label: string;
	
	    static createFrom(source: any = {}) {
	        return new DocLink(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.label = source["label"];
	    }
	}
	export class ScriptMeta {
	    path: string;
	    name: string;
	    desc: string;
	    type: string;
	    language: string;
	    commands: string[];
	    pidFile: string;
	    args: ArgDef[];
	    presets?: CommandPreset[];
	    tags: string[];
	    env?: Record<string, string>;
	    notifyOnDone?: boolean;
	    schedule?: string;
	    requires?: string[];
	    healthUrl?: string;
	    preScripts?: string[];
	    favorite?: boolean;
	    lastRunAt: number;
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new ScriptMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.desc = source["desc"];
	        this.type = source["type"];
	        this.language = source["language"];
	        this.commands = source["commands"];
	        this.pidFile = source["pidFile"];
	        this.args = this.convertValues(source["args"], ArgDef);
	        this.presets = this.convertValues(source["presets"], CommandPreset);
	        this.tags = source["tags"];
	        this.env = source["env"];
	        this.notifyOnDone = source["notifyOnDone"];
	        this.schedule = source["schedule"];
	        this.requires = source["requires"];
	        this.healthUrl = source["healthUrl"];
	        this.preScripts = source["preScripts"];
	        this.favorite = source["favorite"];
	        this.lastRunAt = source["lastRunAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ExportVar {
	    name: string;
	    value: string;
	    kind?: string;
	    secret?: boolean;
	    note?: string;
	
	    static createFrom(source: any = {}) {
	        return new ExportVar(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.value = source["value"];
	        this.kind = source["kind"];
	        this.secret = source["secret"];
	        this.note = source["note"];
	    }
	}
	export class Directory {
	    path: string;
	    isScriptDir: boolean;
	    exportName: string;
	
	    static createFrom(source: any = {}) {
	        return new Directory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.isScriptDir = source["isScriptDir"];
	        this.exportName = source["exportName"];
	    }
	}
	export class Modules {
	    workbench: boolean;
	    ai: boolean;
	    vpn: boolean;
	    memo: boolean;
	    host: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Modules(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workbench = source["workbench"];
	        this.ai = source["ai"];
	        this.vpn = source["vpn"];
	        this.memo = source["memo"];
	        this.host = source["host"];
	    }
	}
	export class Config {
	    modules: Modules;
	    directories: Directory[];
	    exports: ExportVar[];
	    overrideMeta: Record<string, ScriptMeta>;
	    zshrcAnnotations?: Record<string, string>;
	    ai?: AIConfig;
	    autoLaunch?: boolean;
	    traceLogDir?: string;
	    skillsDir?: string;
	    docLinks?: DocLink[];
	    aiProviders?: AIProvider[];
	    activeProviderId?: string;
	    promptTemplates?: Record<string, string>;
	    retention?: DataRetention;
	    memo?: MemoSettings;
	    vpnAppDir?: string;
	    vpnBackupDir?: string;
	    claudeCodePresets?: ClaudeCodePreset[];
	    claudeCodePresetsSeeded?: boolean;
	    codexPresets?: CodexPreset[];
	    codexPresetsSeeded?: boolean;
	    scriptOrders?: Record<string, Array<string>>;
	    mcpServers?: LegacyMCPServer[];
	    mcpMeta?: MCPMeta[];
	    hooks?: HookDef[];
	    hooksSeeded?: boolean;
	    aiWorkspaceProfiles?: AIWorkspaceProfile[];
	    lastAppliedWorkspaceId?: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.modules = this.convertValues(source["modules"], Modules);
	        this.directories = this.convertValues(source["directories"], Directory);
	        this.exports = this.convertValues(source["exports"], ExportVar);
	        this.overrideMeta = this.convertValues(source["overrideMeta"], ScriptMeta, true);
	        this.zshrcAnnotations = source["zshrcAnnotations"];
	        this.ai = this.convertValues(source["ai"], AIConfig);
	        this.autoLaunch = source["autoLaunch"];
	        this.traceLogDir = source["traceLogDir"];
	        this.skillsDir = source["skillsDir"];
	        this.docLinks = this.convertValues(source["docLinks"], DocLink);
	        this.aiProviders = this.convertValues(source["aiProviders"], AIProvider);
	        this.activeProviderId = source["activeProviderId"];
	        this.promptTemplates = source["promptTemplates"];
	        this.retention = this.convertValues(source["retention"], DataRetention);
	        this.memo = this.convertValues(source["memo"], MemoSettings);
	        this.vpnAppDir = source["vpnAppDir"];
	        this.vpnBackupDir = source["vpnBackupDir"];
	        this.claudeCodePresets = this.convertValues(source["claudeCodePresets"], ClaudeCodePreset);
	        this.claudeCodePresetsSeeded = source["claudeCodePresetsSeeded"];
	        this.codexPresets = this.convertValues(source["codexPresets"], CodexPreset);
	        this.codexPresetsSeeded = source["codexPresetsSeeded"];
	        this.scriptOrders = source["scriptOrders"];
	        this.mcpServers = this.convertValues(source["mcpServers"], LegacyMCPServer);
	        this.mcpMeta = this.convertValues(source["mcpMeta"], MCPMeta);
	        this.hooks = this.convertValues(source["hooks"], HookDef);
	        this.hooksSeeded = source["hooksSeeded"];
	        this.aiWorkspaceProfiles = this.convertValues(source["aiWorkspaceProfiles"], AIWorkspaceProfile);
	        this.lastAppliedWorkspaceId = source["lastAppliedWorkspaceId"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	
	
	
	
	
	
	

}

export namespace datadir {
	
	export class Entry {
	    key: string;
	    label: string;
	    path: string;
	    exists: boolean;
	    isDir: boolean;
	    totalBytes: number;
	    fileCount: number;
	    oldestUnix: number;
	    newestUnix: number;
	    cleanable: boolean;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new Entry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.label = source["label"];
	        this.path = source["path"];
	        this.exists = source["exists"];
	        this.isDir = source["isDir"];
	        this.totalBytes = source["totalBytes"];
	        this.fileCount = source["fileCount"];
	        this.oldestUnix = source["oldestUnix"];
	        this.newestUnix = source["newestUnix"];
	        this.cleanable = source["cleanable"];
	        this.description = source["description"];
	    }
	}

}

export namespace diagnostics {
	
	export class HealthCheckResult {
	    script: string;
	    missingDeps: string[];
	    healthUrl: string;
	    healthOk: boolean;
	    healthStatus: number;
	    healthLatencyMs: number;
	    healthError: string;
	    hasAnnotations: boolean;
	
	    static createFrom(source: any = {}) {
	        return new HealthCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.script = source["script"];
	        this.missingDeps = source["missingDeps"];
	        this.healthUrl = source["healthUrl"];
	        this.healthOk = source["healthOk"];
	        this.healthStatus = source["healthStatus"];
	        this.healthLatencyMs = source["healthLatencyMs"];
	        this.healthError = source["healthError"];
	        this.hasAnnotations = source["hasAnnotations"];
	    }
	}
	export class PathEntry {
	    path: string;
	    exists: boolean;
	    duplicate: boolean;
	    shadowsName: string;
	
	    static createFrom(source: any = {}) {
	        return new PathEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.exists = source["exists"];
	        this.duplicate = source["duplicate"];
	        this.shadowsName = source["shadowsName"];
	    }
	}
	export class PathHealth {
	    entries: PathEntry[];
	    missingCount: number;
	    duplicateCount: number;
	
	    static createFrom(source: any = {}) {
	        return new PathHealth(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.entries = this.convertValues(source["entries"], PathEntry);
	        this.missingCount = source["missingCount"];
	        this.duplicateCount = source["duplicateCount"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace executor {
	
	export class RunRecord {
	    id: string;
	    script: string;
	    name: string;
	    args: string[];
	    // Go type: time
	    startedAt: any;
	    // Go type: time
	    endedAt: any;
	    exitCode: number;
	    durationMs: number;
	    stdoutTail: string;
	    logPath: string;
	
	    static createFrom(source: any = {}) {
	        return new RunRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.script = source["script"];
	        this.name = source["name"];
	        this.args = source["args"];
	        this.startedAt = this.convertValues(source["startedAt"], null);
	        this.endedAt = this.convertValues(source["endedAt"], null);
	        this.exitCode = source["exitCode"];
	        this.durationMs = source["durationMs"];
	        this.stdoutTail = source["stdoutTail"];
	        this.logPath = source["logPath"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace fsutil {
	
	export class TextFile {
	    content: string;
	    size: number;
	    binary: boolean;
	    truncated: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TextFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.content = source["content"];
	        this.size = source["size"];
	        this.binary = source["binary"];
	        this.truncated = source["truncated"];
	    }
	}

}

export namespace hooks {
	
	export class EventInfo {
	    event: string;
	    label: string;
	    claude: boolean;
	    codex: boolean;
	    matcherHint: string;
	
	    static createFrom(source: any = {}) {
	        return new EventInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.event = source["event"];
	        this.label = source["label"];
	        this.claude = source["claude"];
	        this.codex = source["codex"];
	        this.matcherHint = source["matcherHint"];
	    }
	}
	export class Stats {
	    total: number;
	    claude: number;
	    codex: number;
	
	    static createFrom(source: any = {}) {
	        return new Stats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.claude = source["claude"];
	        this.codex = source["codex"];
	    }
	}

}

export namespace hostmonitor {
	
	export class DiskBenchmark {
	    bytes: number;
	    writeBytesPerSecond: number;
	    readBytesPerSecond: number;
	
	    static createFrom(source: any = {}) {
	        return new DiskBenchmark(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bytes = source["bytes"];
	        this.writeBytesPerSecond = source["writeBytesPerSecond"];
	        this.readBytesPerSecond = source["readBytesPerSecond"];
	    }
	}
	export class Hardware {
	    model: string;
	    chip: string;
	    cores: string;
	    memory: string;
	    gpu: string;
	    display: string;
	    batteryCycles: string;
	    batteryHealth: string;
	    batteryMaximum: string;
	    os: string;
	
	    static createFrom(source: any = {}) {
	        return new Hardware(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = source["model"];
	        this.chip = source["chip"];
	        this.cores = source["cores"];
	        this.memory = source["memory"];
	        this.gpu = source["gpu"];
	        this.display = source["display"];
	        this.batteryCycles = source["batteryCycles"];
	        this.batteryHealth = source["batteryHealth"];
	        this.batteryMaximum = source["batteryMaximum"];
	        this.os = source["os"];
	    }
	}
	export class NetworkCounter {
	    name: string;
	    received: number;
	    sent: number;
	
	    static createFrom(source: any = {}) {
	        return new NetworkCounter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.received = source["received"];
	        this.sent = source["sent"];
	    }
	}
	export class NetworkQuality {
	    downloadBps: number;
	    uploadBps: number;
	    latencyMs: number;
	    downloadBytes: number;
	    uploadBytes: number;
	    interface: string;
	
	    static createFrom(source: any = {}) {
	        return new NetworkQuality(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.downloadBps = source["downloadBps"];
	        this.uploadBps = source["uploadBps"];
	        this.latencyMs = source["latencyMs"];
	        this.downloadBytes = source["downloadBytes"];
	        this.uploadBytes = source["uploadBytes"];
	        this.interface = source["interface"];
	    }
	}
	export class Port {
	    pid: number;
	    process: string;
	    protocol: string;
	    address: string;
	
	    static createFrom(source: any = {}) {
	        return new Port(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pid = source["pid"];
	        this.process = source["process"];
	        this.protocol = source["protocol"];
	        this.address = source["address"];
	    }
	}
	export class Snapshot {
	    timestamp: number;
	    cpuPercent?: number;
	    cpuTemperature?: number;
	    fanSpeeds: number[];
	    memoryUsed: number;
	    memoryAvailable: number;
	    memoryTotal: number;
	    diskTotal: number;
	    diskAvailable: number;
	    networks: NetworkCounter[];
	    batteryPercent?: number;
	    batteryStatus: string;
	    errors: string[];
	
	    static createFrom(source: any = {}) {
	        return new Snapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.cpuPercent = source["cpuPercent"];
	        this.cpuTemperature = source["cpuTemperature"];
	        this.fanSpeeds = source["fanSpeeds"];
	        this.memoryUsed = source["memoryUsed"];
	        this.memoryAvailable = source["memoryAvailable"];
	        this.memoryTotal = source["memoryTotal"];
	        this.diskTotal = source["diskTotal"];
	        this.diskAvailable = source["diskAvailable"];
	        this.networks = this.convertValues(source["networks"], NetworkCounter);
	        this.batteryPercent = source["batteryPercent"];
	        this.batteryStatus = source["batteryStatus"];
	        this.errors = source["errors"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class StorageEntry {
	    name: string;
	    bytes: number;
	    partial: boolean;
	
	    static createFrom(source: any = {}) {
	        return new StorageEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.bytes = source["bytes"];
	        this.partial = source["partial"];
	    }
	}

}

export namespace leaderboard {
	
	export class Entry {
	    rank: number;
	    model: string;
	    variant?: string;
	    org: string;
	    brand: string;
	    score: number;
	    lower?: number;
	    upper?: number;
	    votes?: number;
	    released?: string;
	
	    static createFrom(source: any = {}) {
	        return new Entry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rank = source["rank"];
	        this.model = source["model"];
	        this.variant = source["variant"];
	        this.org = source["org"];
	        this.brand = source["brand"];
	        this.score = source["score"];
	        this.lower = source["lower"];
	        this.upper = source["upper"];
	        this.votes = source["votes"];
	        this.released = source["released"];
	    }
	}
	export class Board {
	    id: string;
	    domain: string;
	    title: string;
	    tag: string;
	    summary: string;
	    source: string;
	    url: string;
	    license: string;
	    format: string;
	    unit: string;
	    entries: Entry[];
	    total: number;
	    published: string;
	    syncedAt: number;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new Board(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.domain = source["domain"];
	        this.title = source["title"];
	        this.tag = source["tag"];
	        this.summary = source["summary"];
	        this.source = source["source"];
	        this.url = source["url"];
	        this.license = source["license"];
	        this.format = source["format"];
	        this.unit = source["unit"];
	        this.entries = this.convertValues(source["entries"], Entry);
	        this.total = source["total"];
	        this.published = source["published"];
	        this.syncedAt = source["syncedAt"];
	        this.error = source["error"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Snapshot {
	    boards: Board[];
	    refreshing: boolean;
	    stale: boolean;
	    syncedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new Snapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.boards = this.convertValues(source["boards"], Board);
	        this.refreshing = source["refreshing"];
	        this.stale = source["stale"];
	        this.syncedAt = source["syncedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace llm {
	
	export class Trace {
	    id: number;
	    at: string;
	    op: string;
	    model: string;
	    url: string;
	    systemPrompt: string;
	    userPrompt: string;
	    rawResponse: string;
	    error: string;
	    durationMs: number;
	
	    static createFrom(source: any = {}) {
	        return new Trace(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.at = source["at"];
	        this.op = source["op"];
	        this.model = source["model"];
	        this.url = source["url"];
	        this.systemPrompt = source["systemPrompt"];
	        this.userPrompt = source["userPrompt"];
	        this.rawResponse = source["rawResponse"];
	        this.error = source["error"];
	        this.durationMs = source["durationMs"];
	    }
	}
	export class TraceFileInfo {
	    date: string;
	    count: number;
	    bytes: number;
	
	    static createFrom(source: any = {}) {
	        return new TraceFileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.count = source["count"];
	        this.bytes = source["bytes"];
	    }
	}
	export class TraceStorageInfo {
	    dir: string;
	    retainDays: number;
	    files: TraceFileInfo[];
	    totalBytes: number;
	    totalTraces: number;
	
	    static createFrom(source: any = {}) {
	        return new TraceStorageInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dir = source["dir"];
	        this.retainDays = source["retainDays"];
	        this.files = this.convertValues(source["files"], TraceFileInfo);
	        this.totalBytes = source["totalBytes"];
	        this.totalTraces = source["totalTraces"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace mcp {
	
	export class Location {
	    client: string;
	    scope: string;
	    project?: string;
	
	    static createFrom(source: any = {}) {
	        return new Location(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.client = source["client"];
	        this.scope = source["scope"];
	        this.project = source["project"];
	    }
	}
	export class ApplyItem {
	    location: Location;
	    locationKey: string;
	    label: string;
	    file: string;
	    action: string;
	    before?: string;
	    after?: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ApplyItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.location = this.convertValues(source["location"], Location);
	        this.locationKey = source["locationKey"];
	        this.label = source["label"];
	        this.file = source["file"];
	        this.action = source["action"];
	        this.before = source["before"];
	        this.after = source["after"];
	        this.error = source["error"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ApplyPlan {
	    name: string;
	    items: ApplyItem[];
	    canApply: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ApplyPlan(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.items = this.convertValues(source["items"], ApplyItem);
	        this.canApply = source["canApply"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Spec {
	    transport: string;
	    command?: string;
	    args?: string[];
	    env?: Record<string, string>;
	    cwd?: string;
	    url?: string;
	    headers?: Record<string, string>;
	    bearerTokenEnvVar?: string;
	    startupTimeoutSec?: number;
	    toolTimeoutSec?: number;
	    extra?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new Spec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.transport = source["transport"];
	        this.command = source["command"];
	        this.args = source["args"];
	        this.env = source["env"];
	        this.cwd = source["cwd"];
	        this.url = source["url"];
	        this.headers = source["headers"];
	        this.bearerTokenEnvVar = source["bearerTokenEnvVar"];
	        this.startupTimeoutSec = source["startupTimeoutSec"];
	        this.toolTimeoutSec = source["toolTimeoutSec"];
	        this.extra = source["extra"];
	    }
	}
	export class ApplyRequest {
	    name: string;
	    spec: Spec;
	    targets: Location[];
	    sourceClient?: string;
	    origin?: string;
	
	    static createFrom(source: any = {}) {
	        return new ApplyRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.spec = this.convertValues(source["spec"], Spec);
	        this.targets = this.convertValues(source["targets"], Location);
	        this.sourceClient = source["sourceClient"];
	        this.origin = source["origin"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class InputDef {
	    kind: string;
	    key: string;
	    description?: string;
	    required?: boolean;
	    secret?: boolean;
	    default?: string;
	
	    static createFrom(source: any = {}) {
	        return new InputDef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.key = source["key"];
	        this.description = source["description"];
	        this.required = source["required"];
	        this.secret = source["secret"];
	        this.default = source["default"];
	    }
	}
	export class CatalogItem {
	    id: string;
	    title: string;
	    category: string;
	    description: string;
	    registryName?: string;
	    homepage?: string;
	    spec: Spec;
	    inputs?: InputDef[];
	    notes?: string[];
	
	    static createFrom(source: any = {}) {
	        return new CatalogItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.category = source["category"];
	        this.description = source["description"];
	        this.registryName = source["registryName"];
	        this.homepage = source["homepage"];
	        this.spec = this.convertValues(source["spec"], Spec);
	        this.inputs = this.convertValues(source["inputs"], InputDef);
	        this.notes = source["notes"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Draft {
	    name: string;
	    title?: string;
	    description?: string;
	    variant?: string;
	    spec: Spec;
	    inputs?: InputDef[];
	    origin: string;
	    sourceClient?: string;
	    scope?: string;
	    notes?: string[];
	    homepage?: string;
	
	    static createFrom(source: any = {}) {
	        return new Draft(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.variant = source["variant"];
	        this.spec = this.convertValues(source["spec"], Spec);
	        this.inputs = this.convertValues(source["inputs"], InputDef);
	        this.origin = source["origin"];
	        this.sourceClient = source["sourceClient"];
	        this.scope = source["scope"];
	        this.notes = source["notes"];
	        this.homepage = source["homepage"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Entry {
	    name: string;
	    location: Location;
	    locationKey: string;
	    label: string;
	    file?: string;
	    spec: Spec;
	    enabled: boolean;
	    parked?: boolean;
	    stale?: boolean;
	    readOnly?: boolean;
	    readOnlyReason?: string;
	    cliStatus?: string;
	    raw: string;
	    probe?: config.MCPProbe;
	
	    static createFrom(source: any = {}) {
	        return new Entry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.location = this.convertValues(source["location"], Location);
	        this.locationKey = source["locationKey"];
	        this.label = source["label"];
	        this.file = source["file"];
	        this.spec = this.convertValues(source["spec"], Spec);
	        this.enabled = source["enabled"];
	        this.parked = source["parked"];
	        this.stale = source["stale"];
	        this.readOnly = source["readOnly"];
	        this.readOnlyReason = source["readOnlyReason"];
	        this.cliStatus = source["cliStatus"];
	        this.raw = source["raw"];
	        this.probe = this.convertValues(source["probe"], config.MCPProbe);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Source {
	    label: string;
	    path?: string;
	    status: string;
	    error?: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new Source(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.path = source["path"];
	        this.status = source["status"];
	        this.error = source["error"];
	        this.count = source["count"];
	    }
	}
	export class ProjectRef {
	    path: string;
	    name: string;
	    hasServers: boolean;
	    codexTrusted: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProjectRef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.hasServers = source["hasServers"];
	        this.codexTrusted = source["codexTrusted"];
	    }
	}
	export class ServerRow {
	    name: string;
	    entries: Entry[];
	    drift: boolean;
	    meta?: config.MCPMeta;
	
	    static createFrom(source: any = {}) {
	        return new ServerRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.entries = this.convertValues(source["entries"], Entry);
	        this.drift = source["drift"];
	        this.meta = this.convertValues(source["meta"], config.MCPMeta);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Inventory {
	    rows: ServerRow[];
	    projects: ProjectRef[];
	    sources: Source[];
	    cliIncluded: boolean;
	    legacy?: Draft[];
	    scannedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new Inventory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rows = this.convertValues(source["rows"], ServerRow);
	        this.projects = this.convertValues(source["projects"], ProjectRef);
	        this.sources = this.convertValues(source["sources"], Source);
	        this.cliIncluded = source["cliIncluded"];
	        this.legacy = this.convertValues(source["legacy"], Draft);
	        this.scannedAt = source["scannedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class ParseResult {
	    format: string;
	    drafts: Draft[];
	    notes?: string[];
	    registryName?: string;
	
	    static createFrom(source: any = {}) {
	        return new ParseResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.format = source["format"];
	        this.drafts = this.convertValues(source["drafts"], Draft);
	        this.notes = source["notes"];
	        this.registryName = source["registryName"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class RegistryHit {
	    name: string;
	    title?: string;
	    description?: string;
	    version?: string;
	    kinds?: string;
	
	    static createFrom(source: any = {}) {
	        return new RegistryHit(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.version = source["version"];
	        this.kinds = source["kinds"];
	    }
	}
	
	

}

export namespace memo {
	
	export class Attachment {
	    id: string;
	    kind: string;
	    label: string;
	    url?: string;
	    sourcePath?: string;
	    managedPath?: string;
	    mime?: string;
	    size?: number;
	    createdAt: number;
	    cleanedAt?: number;
	
	    static createFrom(source: any = {}) {
	        return new Attachment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.kind = source["kind"];
	        this.label = source["label"];
	        this.url = source["url"];
	        this.sourcePath = source["sourcePath"];
	        this.managedPath = source["managedPath"];
	        this.mime = source["mime"];
	        this.size = source["size"];
	        this.createdAt = source["createdAt"];
	        this.cleanedAt = source["cleanedAt"];
	    }
	}
	export class AttachmentPreview {
	    id: string;
	    kind: string;
	    label: string;
	    dataUrl?: string;
	    path?: string;
	
	    static createFrom(source: any = {}) {
	        return new AttachmentPreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.kind = source["kind"];
	        this.label = source["label"];
	        this.dataUrl = source["dataUrl"];
	        this.path = source["path"];
	    }
	}
	export class Collocation {
	    pattern: string;
	    meaning?: string;
	    example?: string;
	
	    static createFrom(source: any = {}) {
	        return new Collocation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pattern = source["pattern"];
	        this.meaning = source["meaning"];
	        this.example = source["example"];
	    }
	}
	export class Dashboard {
	    today: number;
	    overdue: number;
	    reminders: number;
	    reviewDue: number;
	    totalMemos: number;
	    totalVocab: number;
	
	    static createFrom(source: any = {}) {
	        return new Dashboard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.today = source["today"];
	        this.overdue = source["overdue"];
	        this.reminders = source["reminders"];
	        this.reviewDue = source["reviewDue"];
	        this.totalMemos = source["totalMemos"];
	        this.totalVocab = source["totalVocab"];
	    }
	}
	export class Item {
	    id: string;
	    title: string;
	    body?: string;
	    date: string;
	    status: string;
	    priority?: string;
	    tags?: string[];
	    reminderAt?: number;
	    reminderMode?: string;
	    reminderSentAt?: number;
	    attachments?: Attachment[];
	    createdAt: number;
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new Item(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.body = source["body"];
	        this.date = source["date"];
	        this.status = source["status"];
	        this.priority = source["priority"];
	        this.tags = source["tags"];
	        this.reminderAt = source["reminderAt"];
	        this.reminderMode = source["reminderMode"];
	        this.reminderSentAt = source["reminderSentAt"];
	        this.attachments = this.convertValues(source["attachments"], Attachment);
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LookupResult {
	    word: string;
	    phonetic?: string;
	    audioUrl?: string;
	    partOfSpeech?: string;
	    definitions?: string[];
	    collocations?: Collocation[];
	    source?: string;
	
	    static createFrom(source: any = {}) {
	        return new LookupResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.word = source["word"];
	        this.phonetic = source["phonetic"];
	        this.audioUrl = source["audioUrl"];
	        this.partOfSpeech = source["partOfSpeech"];
	        this.definitions = source["definitions"];
	        this.collocations = this.convertValues(source["collocations"], Collocation);
	        this.source = source["source"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Settings {
	    filesDir: string;
	    retainDays: number;
	    autoCleanup: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filesDir = source["filesDir"];
	        this.retainDays = source["retainDays"];
	        this.autoCleanup = source["autoCleanup"];
	    }
	}
	export class VocabEntry {
	    id: string;
	    word: string;
	    chinese: string;
	    phonetic?: string;
	    partOfSpeech?: string;
	    collocations?: Collocation[];
	    example?: string;
	    audioUrl?: string;
	    tags?: string[];
	    reviewLevel?: number;
	    nextReviewAt?: number;
	    lastReviewedAt?: number;
	    createdAt: number;
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new VocabEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.word = source["word"];
	        this.chinese = source["chinese"];
	        this.phonetic = source["phonetic"];
	        this.partOfSpeech = source["partOfSpeech"];
	        this.collocations = this.convertValues(source["collocations"], Collocation);
	        this.example = source["example"];
	        this.audioUrl = source["audioUrl"];
	        this.tags = source["tags"];
	        this.reviewLevel = source["reviewLevel"];
	        this.nextReviewAt = source["nextReviewAt"];
	        this.lastReviewedAt = source["lastReviewedAt"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace skills {
	
	export class Entry {
	    name: string;
	    description: string;
	    inClaude: boolean;
	    inCodex: boolean;
	    hasAgents: boolean;
	    hasRefs: boolean;
	    modTime: number;
	
	    static createFrom(source: any = {}) {
	        return new Entry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.inClaude = source["inClaude"];
	        this.inCodex = source["inCodex"];
	        this.hasAgents = source["hasAgents"];
	        this.hasRefs = source["hasRefs"];
	        this.modTime = source["modTime"];
	    }
	}
	export class ImportConflict {
	    name: string;
	    source: string;
	    masterModTime: number;
	    sourceModTime: number;
	    masterPath: string;
	    sourcePath: string;
	
	    static createFrom(source: any = {}) {
	        return new ImportConflict(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.source = source["source"];
	        this.masterModTime = source["masterModTime"];
	        this.sourceModTime = source["sourceModTime"];
	        this.masterPath = source["masterPath"];
	        this.sourcePath = source["sourcePath"];
	    }
	}
	export class ImportResult {
	    imported: string[];
	    skipped: string[];
	    conflicts: ImportConflict[];
	
	    static createFrom(source: any = {}) {
	        return new ImportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.imported = source["imported"];
	        this.skipped = source["skipped"];
	        this.conflicts = this.convertValues(source["conflicts"], ImportConflict);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Stats {
	    total: number;
	    claude: number;
	    codex: number;
	    both: number;
	
	    static createFrom(source: any = {}) {
	        return new Stats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.claude = source["claude"];
	        this.codex = source["codex"];
	        this.both = source["both"];
	    }
	}

}

export namespace vpn {
	
	export class BackupPreview {
	    path: string;
	    valid: boolean;
	    numFiles: number;
	    keyFiles: string[];
	    warnings: string[];
	
	    static createFrom(source: any = {}) {
	        return new BackupPreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.valid = source["valid"];
	        this.numFiles = source["numFiles"];
	        this.keyFiles = source["keyFiles"];
	        this.warnings = source["warnings"];
	    }
	}
	export class BackupResult {
	    path: string;
	    size: number;
	    numFiles: number;
	
	    static createFrom(source: any = {}) {
	        return new BackupResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.size = source["size"];
	        this.numFiles = source["numFiles"];
	    }
	}
	export class ConfigFile {
	    name: string;
	    path: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfigFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.content = source["content"];
	    }
	}
	export class CurrentIP {
	    ip: string;
	    country: string;
	    region: string;
	    city: string;
	    org: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new CurrentIP(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ip = source["ip"];
	        this.country = source["country"];
	        this.region = source["region"];
	        this.city = source["city"];
	        this.org = source["org"];
	        this.error = source["error"];
	    }
	}
	export class DNSPolicyEntry {
	    domain: string;
	    server: string;
	    note: string;
	
	    static createFrom(source: any = {}) {
	        return new DNSPolicyEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.domain = source["domain"];
	        this.server = source["server"];
	        this.note = source["note"];
	    }
	}
	export class DiffLine {
	    op: string;
	    text: string;
	
	    static createFrom(source: any = {}) {
	        return new DiffLine(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.op = source["op"];
	        this.text = source["text"];
	    }
	}
	export class EnhanceEditPreview {
	    file: string;
	    kind: string;
	    content: string;
	    diff: DiffLine[];
	    effects: string[];
	    live: boolean;
	    note: string;
	
	    static createFrom(source: any = {}) {
	        return new EnhanceEditPreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file = source["file"];
	        this.kind = source["kind"];
	        this.content = source["content"];
	        this.diff = this.convertValues(source["diff"], DiffLine);
	        this.effects = source["effects"];
	        this.live = source["live"];
	        this.note = source["note"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class EnhanceInfo {
	    file: string;
	    kind: string;
	    scope: string;
	    label: string;
	    desc: string;
	    effect: string;
	    examples: string[];
	
	    static createFrom(source: any = {}) {
	        return new EnhanceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file = source["file"];
	        this.kind = source["kind"];
	        this.scope = source["scope"];
	        this.label = source["label"];
	        this.desc = source["desc"];
	        this.effect = source["effect"];
	        this.examples = source["examples"];
	    }
	}
	export class EnhanceSummary {
	    kind: string;
	    label: string;
	    file: string;
	    customized: boolean;
	    summary: string;
	    desc: string;
	
	    static createFrom(source: any = {}) {
	        return new EnhanceSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.label = source["label"];
	        this.file = source["file"];
	        this.customized = source["customized"];
	        this.summary = source["summary"];
	        this.desc = source["desc"];
	    }
	}
	export class FakeIPEntry {
	    value: string;
	    note: string;
	
	    static createFrom(source: any = {}) {
	        return new FakeIPEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.value = source["value"];
	        this.note = source["note"];
	    }
	}
	export class GroupSummary {
	    name: string;
	    type: string;
	    count: number;
	    now: string;
	
	    static createFrom(source: any = {}) {
	        return new GroupSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.count = source["count"];
	        this.now = source["now"];
	    }
	}
	export class HealthAction {
	    label: string;
	    tab?: string;
	
	    static createFrom(source: any = {}) {
	        return new HealthAction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.tab = source["tab"];
	    }
	}
	export class HealthCheck {
	    id: string;
	    label: string;
	    status: string;
	    detail: string;
	    suggestion?: string;
	
	    static createFrom(source: any = {}) {
	        return new HealthCheck(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.status = source["status"];
	        this.detail = source["detail"];
	        this.suggestion = source["suggestion"];
	    }
	}
	export class HealthReport {
	    overall: string;
	    summary: string;
	    checks: HealthCheck[];
	    actions: HealthAction[];
	    lastBackup?: string;
	
	    static createFrom(source: any = {}) {
	        return new HealthReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.overall = source["overall"];
	        this.summary = source["summary"];
	        this.checks = this.convertValues(source["checks"], HealthCheck);
	        this.actions = this.convertValues(source["actions"], HealthAction);
	        this.lastBackup = source["lastBackup"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class HostEntry {
	    domain: string;
	    value: string;
	    note: string;
	
	    static createFrom(source: any = {}) {
	        return new HostEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.domain = source["domain"];
	        this.value = source["value"];
	        this.note = source["note"];
	    }
	}
	export class PathValidation {
	    appDirOk: boolean;
	    backupDirOk: boolean;
	    errors: string[];
	    warnings: string[];
	
	    static createFrom(source: any = {}) {
	        return new PathValidation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.appDirOk = source["appDirOk"];
	        this.backupDirOk = source["backupDirOk"];
	        this.errors = source["errors"];
	        this.warnings = source["warnings"];
	    }
	}
	export class Profile {
	    uid: string;
	    type: string;
	    name: string;
	    file: string;
	    url?: string;
	    updated: number;
	    isCurrent: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.uid = source["uid"];
	        this.type = source["type"];
	        this.name = source["name"];
	        this.file = source["file"];
	        this.url = source["url"];
	        this.updated = source["updated"];
	        this.isCurrent = source["isCurrent"];
	    }
	}
	export class ProfileFile {
	    file: string;
	    label: string;
	    type: string;
	    remote: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProfileFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file = source["file"];
	        this.label = source["label"];
	        this.type = source["type"];
	        this.remote = source["remote"];
	    }
	}
	export class ProxyGroup {
	    name: string;
	    node: string;
	    delay: number;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.node = source["node"];
	        this.delay = source["delay"];
	        this.type = source["type"];
	    }
	}
	export class RegionCount {
	    name: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new RegionCount(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.count = source["count"];
	    }
	}
	export class RuleTestResult {
	    input: string;
	    expect: string;
	    actual: string;
	    rule: string;
	    pass: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RuleTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.input = source["input"];
	        this.expect = source["expect"];
	        this.actual = source["actual"];
	        this.rule = source["rule"];
	        this.pass = source["pass"];
	    }
	}
	export class RuleChange {
	    section: string;
	    op: string;
	    text: string;
	
	    static createFrom(source: any = {}) {
	        return new RuleChange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.section = source["section"];
	        this.op = source["op"];
	        this.text = source["text"];
	    }
	}
	export class TunEntry {
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new TunEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.value = source["value"];
	    }
	}
	export class RuleItem {
	    type: string;
	    value: string;
	    target: string;
	    noResolve: boolean;
	    raw: string;
	    note: string;
	
	    static createFrom(source: any = {}) {
	        return new RuleItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.value = source["value"];
	        this.target = source["target"];
	        this.noResolve = source["noResolve"];
	        this.raw = source["raw"];
	        this.note = source["note"];
	    }
	}
	export class RulesDraft {
	    prepend: RuleItem[];
	    append: RuleItem[];
	    hosts: HostEntry[];
	    fakeIpFilter: FakeIPEntry[];
	    dnsPolicy: DNSPolicyEntry[];
	    tunExclude: TunEntry[];
	
	    static createFrom(source: any = {}) {
	        return new RulesDraft(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.prepend = this.convertValues(source["prepend"], RuleItem);
	        this.append = this.convertValues(source["append"], RuleItem);
	        this.hosts = this.convertValues(source["hosts"], HostEntry);
	        this.fakeIpFilter = this.convertValues(source["fakeIpFilter"], FakeIPEntry);
	        this.dnsPolicy = this.convertValues(source["dnsPolicy"], DNSPolicyEntry);
	        this.tunExclude = this.convertValues(source["tunExclude"], TunEntry);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RuleAIPreview {
	    summary: string;
	    draft: RulesDraft;
	    changes: RuleChange[];
	    tests: RuleTestResult[];
	
	    static createFrom(source: any = {}) {
	        return new RuleAIPreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.summary = source["summary"];
	        this.draft = this.convertValues(source["draft"], RulesDraft);
	        this.changes = this.convertValues(source["changes"], RuleChange);
	        this.tests = this.convertValues(source["tests"], RuleTestResult);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class RuleIssue {
	    section: string;
	    index: number;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new RuleIssue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.section = source["section"];
	        this.index = source["index"];
	        this.message = source["message"];
	    }
	}
	
	export class RuleMatch {
	    input: string;
	    host: string;
	    ip: string;
	    port: string;
	    hostsValue: string;
	    fakeIpBypass: string;
	    dnsServer: string;
	    tunBypass: string;
	    matched: boolean;
	    rule: string;
	    target: string;
	    source: string;
	    position: number;
	    uncertain: string[];
	    uncertainCount: number;
	
	    static createFrom(source: any = {}) {
	        return new RuleMatch(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.input = source["input"];
	        this.host = source["host"];
	        this.ip = source["ip"];
	        this.port = source["port"];
	        this.hostsValue = source["hostsValue"];
	        this.fakeIpBypass = source["fakeIpBypass"];
	        this.dnsServer = source["dnsServer"];
	        this.tunBypass = source["tunBypass"];
	        this.matched = source["matched"];
	        this.rule = source["rule"];
	        this.target = source["target"];
	        this.source = source["source"];
	        this.position = source["position"];
	        this.uncertain = source["uncertain"];
	        this.uncertainCount = source["uncertainCount"];
	    }
	}
	
	export class RuleTypeInfo {
	    type: string;
	    label: string;
	    desc: string;
	    placeholder: string;
	    noResolve: boolean;
	    advanced: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RuleTypeInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.label = source["label"];
	        this.desc = source["desc"];
	        this.placeholder = source["placeholder"];
	        this.noResolve = source["noResolve"];
	        this.advanced = source["advanced"];
	    }
	}
	
	export class RulesSaveResult {
	    saved: boolean;
	    applied: boolean;
	    message: string;
	    issues: RuleIssue[];
	    warnings: string[];
	
	    static createFrom(source: any = {}) {
	        return new RulesSaveResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.saved = source["saved"];
	        this.applied = source["applied"];
	        this.message = source["message"];
	        this.issues = this.convertValues(source["issues"], RuleIssue);
	        this.warnings = source["warnings"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RulesState {
	    draft: RulesDraft;
	    profileUid: string;
	    profileName: string;
	    rulesFile: string;
	    mergeFile: string;
	    groups: string[];
	    ruleTypes: RuleTypeInfo[];
	    subscriptionRuleCount: number;
	    subscriptionEndsMatch: boolean;
	    subscriptionFakeIp: string[];
	    fakeIpOverridden: boolean;
	    missingFakeIp: string[];
	    deleteCount: number;
	    coreReachable: boolean;
	    tunEnabled: boolean;
	    warnings: string[];
	
	    static createFrom(source: any = {}) {
	        return new RulesState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.draft = this.convertValues(source["draft"], RulesDraft);
	        this.profileUid = source["profileUid"];
	        this.profileName = source["profileName"];
	        this.rulesFile = source["rulesFile"];
	        this.mergeFile = source["mergeFile"];
	        this.groups = source["groups"];
	        this.ruleTypes = this.convertValues(source["ruleTypes"], RuleTypeInfo);
	        this.subscriptionRuleCount = source["subscriptionRuleCount"];
	        this.subscriptionEndsMatch = source["subscriptionEndsMatch"];
	        this.subscriptionFakeIp = source["subscriptionFakeIp"];
	        this.fakeIpOverridden = source["fakeIpOverridden"];
	        this.missingFakeIp = source["missingFakeIp"];
	        this.deleteCount = source["deleteCount"];
	        this.coreReachable = source["coreReachable"];
	        this.tunEnabled = source["tunEnabled"];
	        this.warnings = source["warnings"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Settings {
	    appDir: string;
	    backupDir: string;
	    appDirIsSet: boolean;
	    backupIsSet: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.appDir = source["appDir"];
	        this.backupDir = source["backupDir"];
	        this.appDirIsSet = source["appDirIsSet"];
	        this.backupIsSet = source["backupIsSet"];
	    }
	}
	export class Status {
	    systemProxyEnabled: boolean;
	    tunEnabled: boolean;
	    tunInterface: string;
	    tunExcludedCidrs: string[];
	    currentGroup: string;
	    currentNode: string;
	    proxyGroups: ProxyGroup[];
	    expireAt: number;
	    upload: number;
	    download: number;
	    total: number;
	    coreOnline: boolean;
	    coreVersion: string;
	    mode: string;
	    connections: number;
	    sessionUpload: number;
	    sessionDownload: number;
	    nodeDelay: number;
	
	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.systemProxyEnabled = source["systemProxyEnabled"];
	        this.tunEnabled = source["tunEnabled"];
	        this.tunInterface = source["tunInterface"];
	        this.tunExcludedCidrs = source["tunExcludedCidrs"];
	        this.currentGroup = source["currentGroup"];
	        this.currentNode = source["currentNode"];
	        this.proxyGroups = this.convertValues(source["proxyGroups"], ProxyGroup);
	        this.expireAt = source["expireAt"];
	        this.upload = source["upload"];
	        this.download = source["download"];
	        this.total = source["total"];
	        this.coreOnline = source["coreOnline"];
	        this.coreVersion = source["coreVersion"];
	        this.mode = source["mode"];
	        this.connections = source["connections"];
	        this.sessionUpload = source["sessionUpload"];
	        this.sessionDownload = source["sessionDownload"];
	        this.nodeDelay = source["nodeDelay"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Subscription {
	    uid: string;
	    name: string;
	    type: string;
	    file: string;
	    urlHost: string;
	    isCurrent: boolean;
	    updated: number;
	    upload: number;
	    download: number;
	    total: number;
	    expire: number;
	    updateInterval: number;
	    autoUpdate: boolean;
	    nodeCount: number;
	    ruleCount: number;
	    groups: GroupSummary[];
	    regions: RegionCount[];
	    enhancements: EnhanceSummary[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new Subscription(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.uid = source["uid"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.file = source["file"];
	        this.urlHost = source["urlHost"];
	        this.isCurrent = source["isCurrent"];
	        this.updated = source["updated"];
	        this.upload = source["upload"];
	        this.download = source["download"];
	        this.total = source["total"];
	        this.expire = source["expire"];
	        this.updateInterval = source["updateInterval"];
	        this.autoUpdate = source["autoUpdate"];
	        this.nodeCount = source["nodeCount"];
	        this.ruleCount = source["ruleCount"];
	        this.groups = this.convertValues(source["groups"], GroupSummary);
	        this.regions = this.convertValues(source["regions"], RegionCount);
	        this.enhancements = this.convertValues(source["enhancements"], EnhanceSummary);
	        this.error = source["error"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SubscriptionsView {
	    subscriptions: Subscription[];
	    global: EnhanceSummary[];
	    files: ProfileFile[];
	
	    static createFrom(source: any = {}) {
	        return new SubscriptionsView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.subscriptions = this.convertValues(source["subscriptions"], Subscription);
	        this.global = this.convertValues(source["global"], EnhanceSummary);
	        this.files = this.convertValues(source["files"], ProfileFile);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace workspace {
	
	export class HealthCheck {
	    id: string;
	    label: string;
	    status: string;
	    detail: string;
	    suggestion?: string;
	
	    static createFrom(source: any = {}) {
	        return new HealthCheck(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.status = source["status"];
	        this.detail = source["detail"];
	        this.suggestion = source["suggestion"];
	    }
	}
	export class Health {
	    overall: string;
	    currentProfile: string;
	    profileState: string;
	    checks: HealthCheck[];
	
	    static createFrom(source: any = {}) {
	        return new Health(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.overall = source["overall"];
	        this.currentProfile = source["currentProfile"];
	        this.profileState = source["profileState"];
	        this.checks = this.convertValues(source["checks"], HealthCheck);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Preview {
	    profile: config.AIWorkspaceProfile;
	    changes: string[];
	    errors: string[];
	    warnings: string[];
	    canApply: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Preview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.profile = this.convertValues(source["profile"], config.AIWorkspaceProfile);
	        this.changes = source["changes"];
	        this.errors = source["errors"];
	        this.warnings = source["warnings"];
	        this.canApply = source["canApply"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SnapshotInfo {
	    id: string;
	    profileId: string;
	    profileName: string;
	    createdAt: number;
	
	    static createFrom(source: any = {}) {
	        return new SnapshotInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.profileId = source["profileId"];
	        this.profileName = source["profileName"];
	        this.createdAt = source["createdAt"];
	    }
	}

}

export namespace zshrc {
	
	export class Entry {
	    type: string;
	    key: string;
	    value: string;
	    note: string;
	    lineNum: number;
	
	    static createFrom(source: any = {}) {
	        return new Entry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.key = source["key"];
	        this.value = source["value"];
	        this.note = source["note"];
	        this.lineNum = source["lineNum"];
	    }
	}
	export class Group {
	    kind: string;
	    entries: string[];
	
	    static createFrom(source: any = {}) {
	        return new Group(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.entries = source["entries"];
	    }
	}

}

