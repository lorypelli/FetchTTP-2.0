// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {main} from '../models';

export function CURL(arg1:string):Promise<main.CURLResponse>;

export function CheckUpdates():Promise<main.Update>;

export function HTTP(arg1:string,arg2:string,arg3:Array<main.Header>,arg4:Array<main.Query>,arg5:string):Promise<main.HTTPResponse>;

export function Update():Promise<void>;

export function WS(arg1:string,arg2:Array<main.Header>,arg3:Array<main.Query>,arg4:boolean):Promise<string>;
