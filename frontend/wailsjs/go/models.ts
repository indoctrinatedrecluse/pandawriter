export namespace main {

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
            this.wordErrors = source["wordErrors"] ? source["wordErrors"].map((e: any) => WordError.createFrom(e)) : [];
            this.theme = source["theme"];
            this.font = source["font"];
            this.illustration = source["illustration"];
        }
    }

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