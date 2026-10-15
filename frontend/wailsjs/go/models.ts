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
	export class MCPTestResult {
	    target: string;
	    status: string;
	    summary: string;
	    detail?: string;
	    exitCode?: number;
	    checkedAt?: number;
	    mode?: string;
	
	    static createFrom(source: any = {}) {
	        return new MCPTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.target = source["target"];
	        this.status = source["status"];
	        this.summary = source["summary"];
	        this.detail = source["detail"];
	        this.exitCode = source["exitCode"];
	        this.checkedAt = source["checkedAt"];
	        this.mode = source["mode"];
	    }
	}
	export class MCPServer {
	    id: string;
	    name: string;
	    description?: string;
	    sourceType: string;
	    sourceUrl?: string;
	    localPath?: string;
	    packageName?: string;
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
	    tags?: string[];
	    enabledInClaude?: boolean;
	    enabledInCodex?: boolean;
	    lastTestResults?: Record<string, MCPTestResult>;
	    updatedAt?: number;
	
	    static createFrom(source: any = {}) {
	        return new MCPServer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.sourceType = source["sourceType"];
	        this.sourceUrl = source["sourceUrl"];
	        this.localPath = source["localPath"];
	        this.packageName = source["packageName"];
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
	        this.tags = source["tags"];
	        this.enabledInClaude = source["enabledInClaude"];
	        this.enabledInCodex = source["enabledInCodex"];
	        this.lastTestResults = this.convertValues(source["lastTestResults"], MCPTestResult, true);
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
	
	    static createFrom(source: any = {}) {
	        return new ExportVar(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.value = source["value"];
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
	export class Config {
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
	    mcpDir?: string;
	    mcpServers?: MCPServer[];
	    mcpSeeded?: boolean;
	    hooks?: HookDef[];
	    hooksSeeded?: boolean;
	    aiWorkspaceProfiles?: AIWorkspaceProfile[];
	    lastAppliedWorkspaceId?: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
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
	        this.mcpDir = source["mcpDir"];
	        this.mcpServers = this.convertValues(source["mcpServers"], MCPServer);
	        this.mcpSeeded = source["mcpSeeded"];
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

export namespace main {
	
	export class AICallTrace {
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
	        return new AICallTrace(source);
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
	export class AIEnvironmentHealthCheck {
	    id: string;
	    label: string;
	    status: string;
	    detail: string;
	    suggestion?: string;
	
	    static createFrom(source: any = {}) {
	        return new AIEnvironmentHealthCheck(source);
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
	export class AIEnvironmentHealth {
	    overall: string;
	    currentProfile: string;
	    profileState: string;
	    checks: AIEnvironmentHealthCheck[];
	
	    static createFrom(source: any = {}) {
	        return new AIEnvironmentHealth(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.overall = source["overall"];
	        this.currentProfile = source["currentProfile"];
	        this.profileState = source["profileState"];
	        this.checks = this.convertValues(source["checks"], AIEnvironmentHealthCheck);
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
	export class AITraceFileInfo {
	    date: string;
	    count: number;
	    bytes: number;
	
	    static createFrom(source: any = {}) {
	        return new AITraceFileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.count = source["count"];
	        this.bytes = source["bytes"];
	    }
	}
	export class AITraceStorageInfo {
	    dir: string;
	    retainDays: number;
	    files: AITraceFileInfo[];
	    totalBytes: number;
	    totalTraces: number;
	
	    static createFrom(source: any = {}) {
	        return new AITraceStorageInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dir = source["dir"];
	        this.retainDays = source["retainDays"];
	        this.files = this.convertValues(source["files"], AITraceFileInfo);
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
	export class AIWorkspacePreview {
	    profile: config.AIWorkspaceProfile;
	    changes: string[];
	    errors: string[];
	    warnings: string[];
	    canApply: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AIWorkspacePreview(source);
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
	export class AIWorkspaceSnapshotInfo {
	    id: string;
	    profileId: string;
	    profileName: string;
	    createdAt: number;
	
	    static createFrom(source: any = {}) {
	        return new AIWorkspaceSnapshotInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.profileId = source["profileId"];
	        this.profileName = source["profileName"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class AliasGroup {
	    kind: string;
	    entries: string[];
	
	    static createFrom(source: any = {}) {
	        return new AliasGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.entries = source["entries"];
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
	
	export class DataDirEntry {
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
	        return new DataDirEntry(source);
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
	export class HookEventInfo {
	    event: string;
	    label: string;
	    claude: boolean;
	    codex: boolean;
	    matcherHint: string;
	
	    static createFrom(source: any = {}) {
	        return new HookEventInfo(source);
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
	export class HooksStats {
	    total: number;
	    claude: number;
	    codex: number;
	
	    static createFrom(source: any = {}) {
	        return new HooksStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.claude = source["claude"];
	        this.codex = source["codex"];
	    }
	}
	export class MCPRecommendationCandidate {
	    id: string;
	    name: string;
	    source: string;
	    installMethod: string;
	    version: string;
	    versionNote: string;
	    recommendation: string;
	    fit: string;
	    risk: string;
	    installSteps: string[];
	    server: config.MCPServer;
	    requiredEnv: string[];
	    optionalEnv: string[];
	    quickEnv: Record<string, string>;
	    claudeSupported: boolean;
	    codexSupported: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MCPRecommendationCandidate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.source = source["source"];
	        this.installMethod = source["installMethod"];
	        this.version = source["version"];
	        this.versionNote = source["versionNote"];
	        this.recommendation = source["recommendation"];
	        this.fit = source["fit"];
	        this.risk = source["risk"];
	        this.installSteps = source["installSteps"];
	        this.server = this.convertValues(source["server"], config.MCPServer);
	        this.requiredEnv = source["requiredEnv"];
	        this.optionalEnv = source["optionalEnv"];
	        this.quickEnv = source["quickEnv"];
	        this.claudeSupported = source["claudeSupported"];
	        this.codexSupported = source["codexSupported"];
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
	export class MCPRecommendationResult {
	    query: string;
	    summary: string;
	    sourceNote: string;
	    candidates: MCPRecommendationCandidate[];
	
	    static createFrom(source: any = {}) {
	        return new MCPRecommendationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.query = source["query"];
	        this.summary = source["summary"];
	        this.sourceNote = source["sourceNote"];
	        this.candidates = this.convertValues(source["candidates"], MCPRecommendationCandidate);
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
	export class MCPSettings {
	    mcpDir: string;
	    mcpDirIsSet: boolean;
	    defaultMCPDir: string;
	
	    static createFrom(source: any = {}) {
	        return new MCPSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mcpDir = source["mcpDir"];
	        this.mcpDirIsSet = source["mcpDirIsSet"];
	        this.defaultMCPDir = source["defaultMCPDir"];
	    }
	}
	export class MemoAttachmentPreview {
	    id: string;
	    kind: string;
	    label: string;
	    dataUrl?: string;
	    path?: string;
	
	    static createFrom(source: any = {}) {
	        return new MemoAttachmentPreview(source);
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
	export class MemoDashboard {
	    today: number;
	    overdue: number;
	    reminders: number;
	    reviewDue: number;
	    totalMemos: number;
	    totalVocab: number;
	
	    static createFrom(source: any = {}) {
	        return new MemoDashboard(source);
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
	export class MemoSettingsView {
	    filesDir: string;
	    retainDays: number;
	    autoCleanup: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MemoSettingsView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filesDir = source["filesDir"];
	        this.retainDays = source["retainDays"];
	        this.autoCleanup = source["autoCleanup"];
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
	export class SkillEntry {
	    name: string;
	    description: string;
	    inClaude: boolean;
	    inCodex: boolean;
	    hasAgents: boolean;
	    hasRefs: boolean;
	    modTime: number;
	
	    static createFrom(source: any = {}) {
	        return new SkillEntry(source);
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
	export class SkillImportConflict {
	    name: string;
	    source: string;
	    masterModTime: number;
	    sourceModTime: number;
	    masterPath: string;
	    sourcePath: string;
	
	    static createFrom(source: any = {}) {
	        return new SkillImportConflict(source);
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
	export class SkillImportResult {
	    imported: string[];
	    skipped: string[];
	    conflicts: SkillImportConflict[];
	
	    static createFrom(source: any = {}) {
	        return new SkillImportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.imported = source["imported"];
	        this.skipped = source["skipped"];
	        this.conflicts = this.convertValues(source["conflicts"], SkillImportConflict);
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
	export class SkillsStats {
	    total: number;
	    claude: number;
	    codex: number;
	    both: number;
	
	    static createFrom(source: any = {}) {
	        return new SkillsStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.claude = source["claude"];
	        this.codex = source["codex"];
	        this.both = source["both"];
	    }
	}
	export class VPNBackupPreview {
	    path: string;
	    valid: boolean;
	    numFiles: number;
	    keyFiles: string[];
	    warnings: string[];
	
	    static createFrom(source: any = {}) {
	        return new VPNBackupPreview(source);
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
	export class VPNBackupResult {
	    path: string;
	    size: number;
	    numFiles: number;
	
	    static createFrom(source: any = {}) {
	        return new VPNBackupResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.size = source["size"];
	        this.numFiles = source["numFiles"];
	    }
	}
	export class VPNConfigFile {
	    name: string;
	    path: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new VPNConfigFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.content = source["content"];
	    }
	}
	export class VPNCurrentIP {
	    ip: string;
	    country: string;
	    region: string;
	    city: string;
	    org: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new VPNCurrentIP(source);
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
	export class VPNEnhanceData {
	    prependRules: string[];
	    hosts: Record<string, string>;
	    dnsFakeIPFilter: string[];
	    hasData: boolean;
	    sourceFiles: string[];
	
	    static createFrom(source: any = {}) {
	        return new VPNEnhanceData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.prependRules = source["prependRules"];
	        this.hosts = source["hosts"];
	        this.dnsFakeIPFilter = source["dnsFakeIPFilter"];
	        this.hasData = source["hasData"];
	        this.sourceFiles = source["sourceFiles"];
	    }
	}
	export class VPNHealthAction {
	    label: string;
	    tab?: string;
	
	    static createFrom(source: any = {}) {
	        return new VPNHealthAction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.tab = source["tab"];
	    }
	}
	export class VPNHealthCheck {
	    id: string;
	    label: string;
	    status: string;
	    detail: string;
	    suggestion?: string;
	
	    static createFrom(source: any = {}) {
	        return new VPNHealthCheck(source);
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
	export class VPNHealthReport {
	    overall: string;
	    summary: string;
	    checks: VPNHealthCheck[];
	    actions: VPNHealthAction[];
	    lastBackup?: string;
	
	    static createFrom(source: any = {}) {
	        return new VPNHealthReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.overall = source["overall"];
	        this.summary = source["summary"];
	        this.checks = this.convertValues(source["checks"], VPNHealthCheck);
	        this.actions = this.convertValues(source["actions"], VPNHealthAction);
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
	export class VPNPathValidation {
	    appDirOk: boolean;
	    backupDirOk: boolean;
	    errors: string[];
	    warnings: string[];
	
	    static createFrom(source: any = {}) {
	        return new VPNPathValidation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.appDirOk = source["appDirOk"];
	        this.backupDirOk = source["backupDirOk"];
	        this.errors = source["errors"];
	        this.warnings = source["warnings"];
	    }
	}
	export class VPNProfile {
	    uid: string;
	    type: string;
	    name: string;
	    file: string;
	    url?: string;
	    updated: number;
	    isCurrent: boolean;
	
	    static createFrom(source: any = {}) {
	        return new VPNProfile(source);
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
	export class VPNProxyGroup {
	    name: string;
	    node: string;
	
	    static createFrom(source: any = {}) {
	        return new VPNProxyGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.node = source["node"];
	    }
	}
	export class VPNSettings {
	    appDir: string;
	    backupDir: string;
	    appDirIsSet: boolean;
	    backupIsSet: boolean;
	
	    static createFrom(source: any = {}) {
	        return new VPNSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.appDir = source["appDir"];
	        this.backupDir = source["backupDir"];
	        this.appDirIsSet = source["appDirIsSet"];
	        this.backupIsSet = source["backupIsSet"];
	    }
	}
	export class VPNStatus {
	    systemProxyEnabled: boolean;
	    tunEnabled: boolean;
	    tunInterface: string;
	    tunExcludedCidrs: string[];
	    currentGroup: string;
	    currentNode: string;
	    proxyGroups: VPNProxyGroup[];
	    expireAt: number;
	    upload: number;
	    download: number;
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new VPNStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.systemProxyEnabled = source["systemProxyEnabled"];
	        this.tunEnabled = source["tunEnabled"];
	        this.tunInterface = source["tunInterface"];
	        this.tunExcludedCidrs = source["tunExcludedCidrs"];
	        this.currentGroup = source["currentGroup"];
	        this.currentNode = source["currentNode"];
	        this.proxyGroups = this.convertValues(source["proxyGroups"], VPNProxyGroup);
	        this.expireAt = source["expireAt"];
	        this.upload = source["upload"];
	        this.download = source["download"];
	        this.total = source["total"];
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
	export class VocabLookupResult {
	    word: string;
	    phonetic?: string;
	    audioUrl?: string;
	    partOfSpeech?: string;
	    definitions?: string[];
	    collocations?: memo.Collocation[];
	    source?: string;
	
	    static createFrom(source: any = {}) {
	        return new VocabLookupResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.word = source["word"];
	        this.phonetic = source["phonetic"];
	        this.audioUrl = source["audioUrl"];
	        this.partOfSpeech = source["partOfSpeech"];
	        this.definitions = source["definitions"];
	        this.collocations = this.convertValues(source["collocations"], memo.Collocation);
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

}

