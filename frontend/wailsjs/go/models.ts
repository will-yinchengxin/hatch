export namespace config {
	
	export class AIConfig {
	    enabled: boolean;
	    baseUrl: string;
	    apiKey: string;
	    model: string;
	
	    static createFrom(source: any = {}) {
	        return new AIConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.baseUrl = source["baseUrl"];
	        this.apiKey = source["apiKey"];
	        this.model = source["model"];
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

