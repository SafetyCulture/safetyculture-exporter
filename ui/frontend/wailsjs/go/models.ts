export namespace api {
	
	export class ExporterConfiguration {
	    AccessToken: string;
	    // Go type: struct { ProxyURL string "yaml:\"proxy_url\""; SheqsyURL string "yaml:\"sheqsy_url\""; TLSCert string "yaml:\"tls_cert\""; TLSSkipVerify bool "yaml:\"tls_skip_verify\""; URL string "yaml:\"url\""; MaxConcurrency int "yaml:\"max_concurrency\"" }
	    API: any;
	    // Go type: struct { MaxRowsPerFile int "yaml:\"max_rows_per_file\"" }
	    Csv: any;
	    // Go type: struct { ConnectionString string "yaml:\"connection_string\""; Dialect string "yaml:\"dialect\""; AutoMigrateDisabled bool "yaml:\"auto_migrate_disabled\"" }
	    Db: any;
	    // Go type: struct { Action struct { Limit int "yaml:\"limit\"" } "yaml:\"action\""; Asset struct { Limit int "yaml:\"limit\"" } "yaml:\"asset\""; Course struct { Progress struct { Limit int "yaml:\"limit\""; CompletionStatus string "yaml:\"completion_status\"" } "yaml:\"progress\"" } "yaml:\"course\""; Incremental bool "yaml:\"incremental\""; Inspection struct { Archived string "yaml:\"archived\""; Completed string "yaml:\"completed\""; IncludedInactiveItems bool "yaml:\"included_inactive_items\""; Limit int "yaml:\"limit\""; SkipIds []string "yaml:\"skip_ids\""; WebReportLink string "yaml:\"web_report_link\""; ModifiedBefore api
	    Export: any;
	    // Go type: struct { FilenameConvention string "yaml:\"filename_convention\""; Format []string "yaml:\"format\""; PreferenceID string "yaml:\"preference_id\""; RetryTimeout int "yaml:\"retry_timeout\"" }
	    Report: any;
	    SheqsyCompanyID: string;
	    SheqsyPassword: string;
	    SheqsyUsername: string;
	    // Go type: struct { ExportType string "yaml:\"export_type\"" }
	    Session: any;
	
	    static createFrom(source: any = {}) {
	        return new ExporterConfiguration(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.AccessToken = source["AccessToken"];
	        this.API = this.convertValues(source["API"], Object);
	        this.Csv = this.convertValues(source["Csv"], Object);
	        this.Db = this.convertValues(source["Db"], Object);
	        this.Export = this.convertValues(source["Export"], Object);
	        this.Report = this.convertValues(source["Report"], Object);
	        this.SheqsyCompanyID = source["SheqsyCompanyID"];
	        this.SheqsyPassword = source["SheqsyPassword"];
	        this.SheqsyUsername = source["SheqsyUsername"];
	        this.Session = this.convertValues(source["Session"], Object);
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
	export class TemplateResponseItem {
	    id: string;
	    name: string;
	    // Go type: time
	    modified_at: any;
	
	    static createFrom(source: any = {}) {
	        return new TemplateResponseItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.modified_at = this.convertValues(source["modified_at"], null);
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
	export class mTime {
	
	
	    static createFrom(source: any = {}) {
	        return new mTime(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

export namespace main {
	
	export class VersionResponse {
	    os: string;
	    current: string;
	    latest: string;
	    download_url: string;
	    should_update: boolean;
	
	    static createFrom(source: any = {}) {
	        return new VersionResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.os = source["os"];
	        this.current = source["current"];
	        this.latest = source["latest"];
	        this.download_url = source["download_url"];
	        this.should_update = source["should_update"];
	    }
	}

}

