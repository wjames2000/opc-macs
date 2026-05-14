export namespace agent {
	
	export class AgentUsageItem {
	    agent: string;
	    model: string;
	    calls: number;
	    in_tokens: number;
	    out_tokens: number;
	    cost: number;
	
	    static createFrom(source: any = {}) {
	        return new AgentUsageItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.agent = source["agent"];
	        this.model = source["model"];
	        this.calls = source["calls"];
	        this.in_tokens = source["in_tokens"];
	        this.out_tokens = source["out_tokens"];
	        this.cost = source["cost"];
	    }
	}

}

export namespace main {
	
	export class ActivityLogEntry {
	    agent: string;
	    input: string;
	    output: string;
	    time: string;
	    tokens: number;
	
	    static createFrom(source: any = {}) {
	        return new ActivityLogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.agent = source["agent"];
	        this.input = source["input"];
	        this.output = source["output"];
	        this.time = source["time"];
	        this.tokens = source["tokens"];
	    }
	}
	export class AgentInfo {
	    name: string;
	    summary: string;
	    version: string;
	    model: string;
	    requires_hitl: boolean;
	    tags: string;
	
	    static createFrom(source: any = {}) {
	        return new AgentInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.summary = source["summary"];
	        this.version = source["version"];
	        this.model = source["model"];
	        this.requires_hitl = source["requires_hitl"];
	        this.tags = source["tags"];
	    }
	}
	export class DashboardStats {
	    agent_count: number;
	    total_calls: number;
	    today_calls: number;
	    today_tokens: number;
	    total_cost: string;
	    recent_activity: ActivityLogEntry[];
	
	    static createFrom(source: any = {}) {
	        return new DashboardStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.agent_count = source["agent_count"];
	        this.total_calls = source["total_calls"];
	        this.today_calls = source["today_calls"];
	        this.today_tokens = source["today_tokens"];
	        this.total_cost = source["total_cost"];
	        this.recent_activity = this.convertValues(source["recent_activity"], ActivityLogEntry);
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
	export class ExecuteRequest {
	    agent: string;
	    input: string;
	
	    static createFrom(source: any = {}) {
	        return new ExecuteRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.agent = source["agent"];
	        this.input = source["input"];
	    }
	}
	export class TokenInfo {
	    input: number;
	    output: number;
	
	    static createFrom(source: any = {}) {
	        return new TokenInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.input = source["input"];
	        this.output = source["output"];
	    }
	}
	export class ExecuteResponse {
	    success: boolean;
	    data?: any;
	    error?: string;
	    tokens: TokenInfo;
	
	    static createFrom(source: any = {}) {
	        return new ExecuteResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.data = source["data"];
	        this.error = source["error"];
	        this.tokens = this.convertValues(source["tokens"], TokenInfo);
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
	
	export class UserSettings {
	    api_base_url: string;
	    api_key: string;
	    provider: string;
	    model_name: string;
	    default_agent: string;
	
	    static createFrom(source: any = {}) {
	        return new UserSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.api_base_url = source["api_base_url"];
	        this.api_key = source["api_key"];
	        this.provider = source["provider"];
	        this.model_name = source["model_name"];
	        this.default_agent = source["default_agent"];
	    }
	}

}

