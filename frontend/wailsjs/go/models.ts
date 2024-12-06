export namespace main {
	
	export enum Platform {
	    MacOs = "darwin",
	    Linux = "linux",
	    Windows = "windows",
	}
	export enum TerminalTheme {
	    OneHalf = "OneHalf",
	}
	export enum TerminalFontFamily {
	    FiraCode = "FiraCode",
	}
	export class TerminalFontConfig {
	    Family: TerminalFontFamily;
	    Size: number;
	    Weight: number;
	    WeightBold: number;
	
	    static createFrom(source: any = {}) {
	        return new TerminalFontConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Family = source["Family"];
	        this.Size = source["Size"];
	        this.Weight = source["Weight"];
	        this.WeightBold = source["WeightBold"];
	    }
	}

}

