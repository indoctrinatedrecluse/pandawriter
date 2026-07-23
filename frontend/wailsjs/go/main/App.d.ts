// Generated bindings — updated to match current Go API.
import {main} from '../models';

export function AnalyzeParagraph(arg1:string):Promise<main.Analysis>;

export function CompleteParagraph(arg1:string):Promise<string>;

export function CompleteWord(arg1:string,arg2:string):Promise<string[]>;

export function HasAnyAPIKey():Promise<boolean>;

export function LoadDraft():Promise<main.Draft>;

export function OpenFile():Promise<main.Draft>;

export function SaveDraft(arg1:main.Draft):Promise<void>;

export function SaveFile(arg1:string,arg2:main.Draft):Promise<void>;

export function SaveFileAs(arg1:main.Draft):Promise<string>;