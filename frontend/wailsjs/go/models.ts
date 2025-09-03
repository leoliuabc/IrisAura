export namespace main {
	
	export class CompressRequest {
	    inputDir: string;
	    outputDir: string;
	    format: string;
	    quality: number;
	    maxWidth: number;
	    maxHeight: number;
	
	    static createFrom(source: any = {}) {
	        return new CompressRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.inputDir = source["inputDir"];
	        this.outputDir = source["outputDir"];
	        this.format = source["format"];
	        this.quality = source["quality"];
	        this.maxWidth = source["maxWidth"];
	        this.maxHeight = source["maxHeight"];
	    }
	}
	export class CompressResult {
	    success: boolean;
	    message: string;
	    processedCount: number;
	    totalCount: number;
	    errors: string[];
	    originalSize: number;
	    compressedSize: number;
	
	    static createFrom(source: any = {}) {
	        return new CompressResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.processedCount = source["processedCount"];
	        this.totalCount = source["totalCount"];
	        this.errors = source["errors"];
	        this.originalSize = source["originalSize"];
	        this.compressedSize = source["compressedSize"];
	    }
	}

}

