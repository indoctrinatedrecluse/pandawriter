export namespace main {
	
	export class Draft {
	    exists: boolean;
	    content: string;
	    theme: string;
	    font: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Draft(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.exists = source["exists"];
	        this.content = source["content"];
	        this.theme = source["theme"];
	        this.font = source["font"];
	        this.updatedAt = source["updatedAt"];
	    }
	}

}

