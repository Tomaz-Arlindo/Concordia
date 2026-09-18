export namespace main {
	
	export class ActionResponse {
	    success: boolean;
	    message: string;
	    id?: number;
	
	    static createFrom(source: any = {}) {
	        return new ActionResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.id = source["id"];
	    }
	}
	export class ArticleDetailResponse {
	    success: boolean;
	    message: string;
	    article?: models.ArticleDetail;
	
	    static createFrom(source: any = {}) {
	        return new ArticleDetailResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.article = this.convertValues(source["article"], models.ArticleDetail);
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
	export class AuthResponse {
	    success: boolean;
	    message: string;
	    user?: models.User;
	
	    static createFrom(source: any = {}) {
	        return new AuthResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.user = this.convertValues(source["user"], models.User);
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
	export class BatchImportResponse {
	    success: boolean;
	    message: string;
	    selected: number;
	    imported: number;
	    duplicates: number;
	    failed: number;
	    errors: string[];
	
	    static createFrom(source: any = {}) {
	        return new BatchImportResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.selected = source["selected"];
	        this.imported = source["imported"];
	        this.duplicates = source["duplicates"];
	        this.failed = source["failed"];
	        this.errors = source["errors"];
	    }
	}
	export class BenchmarkResponse {
	    success: boolean;
	    message: string;
	    report: models.BenchmarkReport;
	
	    static createFrom(source: any = {}) {
	        return new BenchmarkResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.report = this.convertValues(source["report"], models.BenchmarkReport);
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
	export class FileSelectionResponse {
	    success: boolean;
	    message: string;
	    name?: string;
	    size?: number;
	
	    static createFrom(source: any = {}) {
	        return new FileSelectionResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.name = source["name"];
	        this.size = source["size"];
	    }
	}
	export class HealthResponse {
	    status: string;
	    database: string;
	
	    static createFrom(source: any = {}) {
	        return new HealthResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.database = source["database"];
	    }
	}
	export class IndexingResponse {
	    success: boolean;
	    message: string;
	    report: models.IndexingReport;
	
	    static createFrom(source: any = {}) {
	        return new IndexingResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.report = this.convertValues(source["report"], models.IndexingReport);
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

export namespace models {
	
	export class ArticleDetail {
	    id: number;
	    title: string;
	    description: string;
	    filename: string;
	    file_path: string;
	    content_type: string;
	    size: number;
	    user_id: number;
	    source_id: number;
	    source_name: string;
	    authors: string;
	    abstract_text: string;
	    external_id: string;
	    doi: string;
	    language_code: string;
	    publication_date: string;
	    checksum_sha256: string;
	    index_status: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new ArticleDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.filename = source["filename"];
	        this.file_path = source["file_path"];
	        this.content_type = source["content_type"];
	        this.size = source["size"];
	        this.user_id = source["user_id"];
	        this.source_id = source["source_id"];
	        this.source_name = source["source_name"];
	        this.authors = source["authors"];
	        this.abstract_text = source["abstract_text"];
	        this.external_id = source["external_id"];
	        this.doi = source["doi"];
	        this.language_code = source["language_code"];
	        this.publication_date = source["publication_date"];
	        this.checksum_sha256 = source["checksum_sha256"];
	        this.index_status = source["index_status"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
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
	export class ArticleSource {
	    id: number;
	    name: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new ArticleSource(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	    }
	}
	export class ArticleSummary {
	    id: number;
	    title: string;
	    description: string;
	    filename: string;
	    source_name: string;
	    authors: string;
	    language_code: string;
	    publication_date: string;
	    index_status: string;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new ArticleSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.filename = source["filename"];
	        this.source_name = source["source_name"];
	        this.authors = source["authors"];
	        this.language_code = source["language_code"];
	        this.publication_date = source["publication_date"];
	        this.index_status = source["index_status"];
	        this.created_at = this.convertValues(source["created_at"], null);
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
	export class BenchmarkResult {
	    workers: number;
	    processing_ms: number;
	    database_ms: number;
	    duration_ms: number;
	    indexed: number;
	    failed: number;
	    processing_speedup: number;
	    processing_efficiency: number;
	    total_speedup: number;
	    database_share: number;
	    errors: string[];
	
	    static createFrom(source: any = {}) {
	        return new BenchmarkResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workers = source["workers"];
	        this.processing_ms = source["processing_ms"];
	        this.database_ms = source["database_ms"];
	        this.duration_ms = source["duration_ms"];
	        this.indexed = source["indexed"];
	        this.failed = source["failed"];
	        this.processing_speedup = source["processing_speedup"];
	        this.processing_efficiency = source["processing_efficiency"];
	        this.total_speedup = source["total_speedup"];
	        this.database_share = source["database_share"];
	        this.errors = source["errors"];
	    }
	}
	export class BenchmarkReport {
	    results: BenchmarkResult[];
	    note: string;
	
	    static createFrom(source: any = {}) {
	        return new BenchmarkReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.results = this.convertValues(source["results"], BenchmarkResult);
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
	
	export class CollectionStats {
	    total_articles: number;
	    pending_articles: number;
	    indexed_articles: number;
	    total_authors: number;
	    total_sources: number;
	
	    static createFrom(source: any = {}) {
	        return new CollectionStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_articles = source["total_articles"];
	        this.pending_articles = source["pending_articles"];
	        this.indexed_articles = source["indexed_articles"];
	        this.total_authors = source["total_authors"];
	        this.total_sources = source["total_sources"];
	    }
	}
	export class IndexingOverview {
	    pending: number;
	    processing: number;
	    indexed: number;
	    errors: number;
	    terms: number;
	    occurrences: number;
	    documents: number;
	
	    static createFrom(source: any = {}) {
	        return new IndexingOverview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pending = source["pending"];
	        this.processing = source["processing"];
	        this.indexed = source["indexed"];
	        this.errors = source["errors"];
	        this.terms = source["terms"];
	        this.occurrences = source["occurrences"];
	        this.documents = source["documents"];
	    }
	}
	export class IndexingReport {
	    workers: number;
	    total: number;
	    indexed: number;
	    failed: number;
	    duration_ms: number;
	    reindexed_all: boolean;
	    errors: string[];
	
	    static createFrom(source: any = {}) {
	        return new IndexingReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workers = source["workers"];
	        this.total = source["total"];
	        this.indexed = source["indexed"];
	        this.failed = source["failed"];
	        this.duration_ms = source["duration_ms"];
	        this.reindexed_all = source["reindexed_all"];
	        this.errors = source["errors"];
	    }
	}
	export class SearchResult {
	    id: number;
	    title: string;
	    description: string;
	    filename: string;
	    source_name: string;
	    authors: string;
	    language_code: string;
	    publication_date: string;
	    index_status: string;
	    score: number;
	    matched_terms: string;
	
	    static createFrom(source: any = {}) {
	        return new SearchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.filename = source["filename"];
	        this.source_name = source["source_name"];
	        this.authors = source["authors"];
	        this.language_code = source["language_code"];
	        this.publication_date = source["publication_date"];
	        this.index_status = source["index_status"];
	        this.score = source["score"];
	        this.matched_terms = source["matched_terms"];
	    }
	}
	export class User {
	    id: number;
	    name: string;
	    email: string;
	    role: string;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new User(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.email = source["email"];
	        this.role = source["role"];
	        this.created_at = this.convertValues(source["created_at"], null);
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

