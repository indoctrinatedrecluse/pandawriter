export namespace main {
	
	export class WordError {
	    incorrect: string;
	    correct: string;
	
	    static createFrom(source: any = {}) {
	        return new WordError(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.incorrect = source["incorrect"];
	        this.correct = source["correct"];
	    }
	}
	export class Analysis {
	    wordErrors: WordError[];
	    theme: string;
	    font: string;
	    illustration: string;
	
	    static createFrom(source: any = {}) {
	        return new Analysis(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.wordErrors = this.convertValues(source["wordErrors"], WordError);
	        this.theme = source["theme"];
	        this.font = source["font"];
	        this.illustration = source["illustration"];
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
	    exists: boolean;
	    title: string;
	    content: string;
	    theme: string;
	    font: string;
	    fontSize: string;
	    spacing: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Draft(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.exists = source["exists"];
	        this.title = source["title"];
	        this.content = source["content"];
	        this.theme = source["theme"];
	        this.font = source["font"];
	        this.fontSize = source["fontSize"];
	        this.spacing = source["spacing"];
	        this.updatedAt = source["updatedAt"];
	    }
	}

}

