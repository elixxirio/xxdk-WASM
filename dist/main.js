/*
 * ATTENTION: The "eval" devtool has been used (maybe by default in mode: "development").
 * This devtool is neither made for production nor for readable output files.
 * It uses "eval()" calls to create a separate source file in the browser devtools.
 * If you are trying to read the output file, select a different devtool (https://webpack.js.org/configuration/devtool/)
 * or disable the default devtool with "devtool: false".
 * If you are looking for production-ready output files, see mode: "production" (https://webpack.js.org/configuration/mode/).
 */
/******/ (() => { // webpackBootstrap
/******/ 	"use strict";
/******/ 	var __webpack_modules__ = ({

/***/ "./node_modules/golang-wasm/src/bridge.js":
/*!************************************************!*\
  !*** ./node_modules/golang-wasm/src/bridge.js ***!
  \************************************************/
/***/ ((__unused_webpack_module, __webpack_exports__, __webpack_require__) => {

eval("__webpack_require__.r(__webpack_exports__);\n/* harmony export */ __webpack_require__.d(__webpack_exports__, {\n/* harmony export */   \"default\": () => (/* export default binding */ __WEBPACK_DEFAULT_EXPORT__)\n/* harmony export */ });\n// Initially, the __go_wasm__ object will be an empty object. \nconst g = __webpack_require__.g || window || self;\nif (!g.__go_wasm__) {\n    g.__go_wasm__ = {};\n}\n\n/**\n * The maximum amount of time that we would expect Wasm to take to initialize.\n * If it doesn't initialize after this time, we send a warning to console.\n * Most likely something has gone wrong if it takes more than 3 seconds to initialize.\n */\nconst maxTime = 3 * 1000;\n\n/**\n * bridge is an easier way to refer to the Go WASM object.\n */\nconst bridge = g.__go_wasm__;\n\n/**\n * Wrapper is used by Go to run all Go functions in JS.\n * \n * @param {Function} goFunc a function that is expected to return an object of the following specification:\n * {\n *  result:  undefined | any         // undefined when error is returned, or function returns undefined.\n *  error:       Error | undefined   // undefined when no error is present.\n * }\n * \n * @returns {Function} returns a function that take arguments which are used to call the Go function.\n */\nfunction wrapper(goFunc) {\n    return (...args) => {\n        const result = goFunc.apply(undefined, args);\n        if (result.error instanceof Error) {\n            throw result.error;\n        }\n        return result.result;\n    }\n}\n\n/**\n * Sleep is used when awaiting for Go Wasm to initialize.\n * It uses the lowest possible sane delay time (via requestAnimationFrame).\n * However, if the window is not focused, requestAnimationFrame never returns.\n * A timeout will ensure to be called after 50 ms, regardless of whether or not the tab is in focus.\n * \n * @returns {Promise} an always-resolving promise when a tick has been completed.\n */\nfunction sleep() {\n    return new Promise((res) => {\n        requestAnimationFrame(() => res());\n        setTimeout(() => {\n            res();\n        }, 50);\n    });\n}\n\n/**\n * @param {ArrayBuffer} getBytes a promise that is bytes of the Go Wasm object.\n * \n * @returns {Proxy} an object that can be used to call WASM's objects and properly parse their results.\n * \n * All values that want to be retrieved from the proxy, regardless of if they are a function or not, must be retrieved\n * as if they are from a function call. \n * \n * If a non-function value is returned however arguments are provided, a warning will be printed.\n */\n/* harmony default export */ function __WEBPACK_DEFAULT_EXPORT__(getBytes) {\n    let proxy;\n    let go;\n\n    async function init() {\n        bridge.__wrapper__ = wrapper;\n\n        go = new g.Go();\n        let bytes = await getBytes;\n        let result = await WebAssembly.instantiate(bytes, go.importObject);\n        go.run(result.instance);\n    }\n\n    init();\n    setTimeout(() => {\n        if (bridge.__ready__ !== true) {\n            console.warn(\"Golang WASM Bridge (__go_wasm__.__ready__) still not true after max time\");\n        }\n    }, maxTime);\n\n\n    proxy = new Proxy(\n        {},\n        {\n            get: (_, key) => {\n                return (...args) => {\n                    return new Promise(async (res, rej) => {\n                        if (!go || go.exited) {\n                            return rej(new Error(\"The Go instance is not active.\"));\n                        }\n                        while (bridge.__ready__ !== true) {\n                            await sleep();\n                        }\n\n                        if (typeof bridge[key] !== 'function') {\n                            res(bridge[key]);\n\n                            if (args.length !== 0) {\n                                console.warn(\"Retrieved value from WASM returned function type, however called with arguments.\")\n                            }\n                            return;\n                        }\n\n                        try {\n                            res(bridge[key].apply(undefined, args));\n                        } catch (e) {\n                            rej(e);\n                        }\n                    })\n                };\n            }\n        }\n    );\n\n    bridge.__proxy__ = proxy;\n    return proxy;\n}\n\n\n//# sourceURL=webpack://xxdk-wasm/./node_modules/golang-wasm/src/bridge.js?");

/***/ }),

/***/ "./src/api/main.go":
/*!*************************!*\
  !*** ./src/api/main.go ***!
  \*************************/
/***/ ((__unused_webpack_module, __webpack_exports__, __webpack_require__) => {

eval("__webpack_require__.r(__webpack_exports__);\n/* harmony export */ __webpack_require__.d(__webpack_exports__, {\n/* harmony export */   \"default\": () => (__WEBPACK_DEFAULT_EXPORT__)\n/* harmony export */ });\n/* harmony import */ var C_Users_Jono_Projects_Go_git_xx_network_elixxir_xxdk_wasm2_node_modules_golang_wasm_src_bridge_js__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! ./node_modules/golang-wasm/src/bridge.js */ \"./node_modules/golang-wasm/src/bridge.js\");\n__webpack_require__(/*! !./node_modules/golang-wasm/src/wasm_exec.js */ \"./node_modules/golang-wasm/src/wasm_exec.js\");\n\n\nconst wasm = fetch('main.wasm').then(response => response.arrayBuffer());\n/* harmony default export */ const __WEBPACK_DEFAULT_EXPORT__ = ((0,C_Users_Jono_Projects_Go_git_xx_network_elixxir_xxdk_wasm2_node_modules_golang_wasm_src_bridge_js__WEBPACK_IMPORTED_MODULE_0__[\"default\"])(wasm));\n\n//# sourceURL=webpack://xxdk-wasm/./src/api/main.go?");

/***/ }),

/***/ "./node_modules/golang-wasm/src/wasm_exec.js":
/*!***************************************************!*\
  !*** ./node_modules/golang-wasm/src/wasm_exec.js ***!
  \***************************************************/
/***/ (() => {

eval("// Copyright 2018 The Go Authors. All rights reserved.\n// Use of this source code is governed by a BSD-style\n// license that can be found in the LICENSE file.\n\n\n\n(() => {\n\tconst enosys = () => {\n\t\tconst err = new Error(\"not implemented\");\n\t\terr.code = \"ENOSYS\";\n\t\treturn err;\n\t};\n\n\tif (!globalThis.fs) {\n\t\tlet outputBuf = \"\";\n\t\tglobalThis.fs = {\n\t\t\tconstants: { O_WRONLY: -1, O_RDWR: -1, O_CREAT: -1, O_TRUNC: -1, O_APPEND: -1, O_EXCL: -1 }, // unused\n\t\t\twriteSync(fd, buf) {\n\t\t\t\toutputBuf += decoder.decode(buf);\n\t\t\t\tconst nl = outputBuf.lastIndexOf(\"\\n\");\n\t\t\t\tif (nl != -1) {\n\t\t\t\t\tconsole.log(outputBuf.substr(0, nl));\n\t\t\t\t\toutputBuf = outputBuf.substr(nl + 1);\n\t\t\t\t}\n\t\t\t\treturn buf.length;\n\t\t\t},\n\t\t\twrite(fd, buf, offset, length, position, callback) {\n\t\t\t\tif (offset !== 0 || length !== buf.length || position !== null) {\n\t\t\t\t\tcallback(enosys());\n\t\t\t\t\treturn;\n\t\t\t\t}\n\t\t\t\tconst n = this.writeSync(fd, buf);\n\t\t\t\tcallback(null, n);\n\t\t\t},\n\t\t\tchmod(path, mode, callback) { callback(enosys()); },\n\t\t\tchown(path, uid, gid, callback) { callback(enosys()); },\n\t\t\tclose(fd, callback) { callback(enosys()); },\n\t\t\tfchmod(fd, mode, callback) { callback(enosys()); },\n\t\t\tfchown(fd, uid, gid, callback) { callback(enosys()); },\n\t\t\tfstat(fd, callback) { callback(enosys()); },\n\t\t\tfsync(fd, callback) { callback(null); },\n\t\t\tftruncate(fd, length, callback) { callback(enosys()); },\n\t\t\tlchown(path, uid, gid, callback) { callback(enosys()); },\n\t\t\tlink(path, link, callback) { callback(enosys()); },\n\t\t\tlstat(path, callback) { callback(enosys()); },\n\t\t\tmkdir(path, perm, callback) { callback(enosys()); },\n\t\t\topen(path, flags, mode, callback) { callback(enosys()); },\n\t\t\tread(fd, buffer, offset, length, position, callback) { callback(enosys()); },\n\t\t\treaddir(path, callback) { callback(enosys()); },\n\t\t\treadlink(path, callback) { callback(enosys()); },\n\t\t\trename(from, to, callback) { callback(enosys()); },\n\t\t\trmdir(path, callback) { callback(enosys()); },\n\t\t\tstat(path, callback) { callback(enosys()); },\n\t\t\tsymlink(path, link, callback) { callback(enosys()); },\n\t\t\ttruncate(path, length, callback) { callback(enosys()); },\n\t\t\tunlink(path, callback) { callback(enosys()); },\n\t\t\tutimes(path, atime, mtime, callback) { callback(enosys()); },\n\t\t};\n\t}\n\n\tif (!globalThis.process) {\n\t\tglobalThis.process = {\n\t\t\tgetuid() { return -1; },\n\t\t\tgetgid() { return -1; },\n\t\t\tgeteuid() { return -1; },\n\t\t\tgetegid() { return -1; },\n\t\t\tgetgroups() { throw enosys(); },\n\t\t\tpid: -1,\n\t\t\tppid: -1,\n\t\t\tumask() { throw enosys(); },\n\t\t\tcwd() { throw enosys(); },\n\t\t\tchdir() { throw enosys(); },\n\t\t}\n\t}\n\n\tif (!globalThis.crypto) {\n\t\tthrow new Error(\"globalThis.crypto is not available, polyfill required (crypto.getRandomValues only)\");\n\t}\n\n\tif (!globalThis.performance) {\n\t\tthrow new Error(\"globalThis.performance is not available, polyfill required (performance.now only)\");\n\t}\n\n\tif (!globalThis.TextEncoder) {\n\t\tthrow new Error(\"globalThis.TextEncoder is not available, polyfill required\");\n\t}\n\n\tif (!globalThis.TextDecoder) {\n\t\tthrow new Error(\"globalThis.TextDecoder is not available, polyfill required\");\n\t}\n\n\tconst encoder = new TextEncoder(\"utf-8\");\n\tconst decoder = new TextDecoder(\"utf-8\");\n\n\tglobalThis.Go = class {\n\t\tconstructor() {\n\t\t\tthis.argv = [\"js\"];\n\t\t\tthis.env = {};\n\t\t\tthis.exit = (code) => {\n\t\t\t\tif (code !== 0) {\n\t\t\t\t\tconsole.warn(\"exit code:\", code);\n\t\t\t\t}\n\t\t\t};\n\t\t\tthis._exitPromise = new Promise((resolve) => {\n\t\t\t\tthis._resolveExitPromise = resolve;\n\t\t\t});\n\t\t\tthis._pendingEvent = null;\n\t\t\tthis._scheduledTimeouts = new Map();\n\t\t\tthis._nextCallbackTimeoutID = 1;\n\n\t\t\tconst setInt64 = (addr, v) => {\n\t\t\t\tthis.mem.setUint32(addr + 0, v, true);\n\t\t\t\tthis.mem.setUint32(addr + 4, Math.floor(v / 4294967296), true);\n\t\t\t}\n\n\t\t\tconst getInt64 = (addr) => {\n\t\t\t\tconst low = this.mem.getUint32(addr + 0, true);\n\t\t\t\tconst high = this.mem.getInt32(addr + 4, true);\n\t\t\t\treturn low + high * 4294967296;\n\t\t\t}\n\n\t\t\tconst loadValue = (addr) => {\n\t\t\t\tconst f = this.mem.getFloat64(addr, true);\n\t\t\t\tif (f === 0) {\n\t\t\t\t\treturn undefined;\n\t\t\t\t}\n\t\t\t\tif (!isNaN(f)) {\n\t\t\t\t\treturn f;\n\t\t\t\t}\n\n\t\t\t\tconst id = this.mem.getUint32(addr, true);\n\t\t\t\treturn this._values[id];\n\t\t\t}\n\n\t\t\tconst storeValue = (addr, v) => {\n\t\t\t\tconst nanHead = 0x7FF80000;\n\n\t\t\t\tif (typeof v === \"number\" && v !== 0) {\n\t\t\t\t\tif (isNaN(v)) {\n\t\t\t\t\t\tthis.mem.setUint32(addr + 4, nanHead, true);\n\t\t\t\t\t\tthis.mem.setUint32(addr, 0, true);\n\t\t\t\t\t\treturn;\n\t\t\t\t\t}\n\t\t\t\t\tthis.mem.setFloat64(addr, v, true);\n\t\t\t\t\treturn;\n\t\t\t\t}\n\n\t\t\t\tif (v === undefined) {\n\t\t\t\t\tthis.mem.setFloat64(addr, 0, true);\n\t\t\t\t\treturn;\n\t\t\t\t}\n\n\t\t\t\tlet id = this._ids.get(v);\n\t\t\t\tif (id === undefined) {\n\t\t\t\t\tid = this._idPool.pop();\n\t\t\t\t\tif (id === undefined) {\n\t\t\t\t\t\tid = this._values.length;\n\t\t\t\t\t}\n\t\t\t\t\tthis._values[id] = v;\n\t\t\t\t\tthis._goRefCounts[id] = 0;\n\t\t\t\t\tthis._ids.set(v, id);\n\t\t\t\t}\n\t\t\t\tthis._goRefCounts[id]++;\n\t\t\t\tlet typeFlag = 0;\n\t\t\t\tswitch (typeof v) {\n\t\t\t\t\tcase \"object\":\n\t\t\t\t\t\tif (v !== null) {\n\t\t\t\t\t\t\ttypeFlag = 1;\n\t\t\t\t\t\t}\n\t\t\t\t\t\tbreak;\n\t\t\t\t\tcase \"string\":\n\t\t\t\t\t\ttypeFlag = 2;\n\t\t\t\t\t\tbreak;\n\t\t\t\t\tcase \"symbol\":\n\t\t\t\t\t\ttypeFlag = 3;\n\t\t\t\t\t\tbreak;\n\t\t\t\t\tcase \"function\":\n\t\t\t\t\t\ttypeFlag = 4;\n\t\t\t\t\t\tbreak;\n\t\t\t\t}\n\t\t\t\tthis.mem.setUint32(addr + 4, nanHead | typeFlag, true);\n\t\t\t\tthis.mem.setUint32(addr, id, true);\n\t\t\t}\n\n\t\t\tconst loadSlice = (addr) => {\n\t\t\t\tconst array = getInt64(addr + 0);\n\t\t\t\tconst len = getInt64(addr + 8);\n\t\t\t\treturn new Uint8Array(this._inst.exports.mem.buffer, array, len);\n\t\t\t}\n\n\t\t\tconst loadSliceOfValues = (addr) => {\n\t\t\t\tconst array = getInt64(addr + 0);\n\t\t\t\tconst len = getInt64(addr + 8);\n\t\t\t\tconst a = new Array(len);\n\t\t\t\tfor (let i = 0; i < len; i++) {\n\t\t\t\t\ta[i] = loadValue(array + i * 8);\n\t\t\t\t}\n\t\t\t\treturn a;\n\t\t\t}\n\n\t\t\tconst loadString = (addr) => {\n\t\t\t\tconst saddr = getInt64(addr + 0);\n\t\t\t\tconst len = getInt64(addr + 8);\n\t\t\t\treturn decoder.decode(new DataView(this._inst.exports.mem.buffer, saddr, len));\n\t\t\t}\n\n\t\t\tconst timeOrigin = Date.now() - performance.now();\n\t\t\tthis.importObject = {\n\t\t\t\tgo: {\n\t\t\t\t\t// Go's SP does not change as long as no Go code is running. Some operations (e.g. calls, getters and setters)\n\t\t\t\t\t// may synchronously trigger a Go event handler. This makes Go code get executed in the middle of the imported\n\t\t\t\t\t// function. A goroutine can switch to a new stack if the current stack is too small (see morestack function).\n\t\t\t\t\t// This changes the SP, thus we have to update the SP used by the imported function.\n\n\t\t\t\t\t// func wasmExit(code int32)\n\t\t\t\t\t\"runtime.wasmExit\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\tconst code = this.mem.getInt32(sp + 8, true);\n\t\t\t\t\t\tthis.exited = true;\n\t\t\t\t\t\tdelete this._inst;\n\t\t\t\t\t\tdelete this._values;\n\t\t\t\t\t\tdelete this._goRefCounts;\n\t\t\t\t\t\tdelete this._ids;\n\t\t\t\t\t\tdelete this._idPool;\n\t\t\t\t\t\tthis.exit(code);\n\t\t\t\t\t},\n\n\t\t\t\t\t// func wasmWrite(fd uintptr, p unsafe.Pointer, n int32)\n\t\t\t\t\t\"runtime.wasmWrite\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\tconst fd = getInt64(sp + 8);\n\t\t\t\t\t\tconst p = getInt64(sp + 16);\n\t\t\t\t\t\tconst n = this.mem.getInt32(sp + 24, true);\n\t\t\t\t\t\tfs.writeSync(fd, new Uint8Array(this._inst.exports.mem.buffer, p, n));\n\t\t\t\t\t},\n\n\t\t\t\t\t// func resetMemoryDataView()\n\t\t\t\t\t\"runtime.resetMemoryDataView\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\tthis.mem = new DataView(this._inst.exports.mem.buffer);\n\t\t\t\t\t},\n\n\t\t\t\t\t// func nanotime1() int64\n\t\t\t\t\t\"runtime.nanotime1\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\tsetInt64(sp + 8, (timeOrigin + performance.now()) * 1000000);\n\t\t\t\t\t},\n\n\t\t\t\t\t// func walltime() (sec int64, nsec int32)\n\t\t\t\t\t\"runtime.walltime\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\tconst msec = (new Date).getTime();\n\t\t\t\t\t\tsetInt64(sp + 8, msec / 1000);\n\t\t\t\t\t\tthis.mem.setInt32(sp + 16, (msec % 1000) * 1000000, true);\n\t\t\t\t\t},\n\n\t\t\t\t\t// func scheduleTimeoutEvent(delay int64) int32\n\t\t\t\t\t\"runtime.scheduleTimeoutEvent\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\tconst id = this._nextCallbackTimeoutID;\n\t\t\t\t\t\tthis._nextCallbackTimeoutID++;\n\t\t\t\t\t\tthis._scheduledTimeouts.set(id, setTimeout(\n\t\t\t\t\t\t\t() => {\n\t\t\t\t\t\t\t\tthis._resume();\n\t\t\t\t\t\t\t\twhile (this._scheduledTimeouts.has(id)) {\n\t\t\t\t\t\t\t\t\t// for some reason Go failed to register the timeout event, log and try again\n\t\t\t\t\t\t\t\t\t// (temporary workaround for https://github.com/golang/go/issues/28975)\n\t\t\t\t\t\t\t\t\tconsole.warn(\"scheduleTimeoutEvent: missed timeout event\");\n\t\t\t\t\t\t\t\t\tthis._resume();\n\t\t\t\t\t\t\t\t}\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\tgetInt64(sp + 8) + 1, // setTimeout has been seen to fire up to 1 millisecond early\n\t\t\t\t\t\t));\n\t\t\t\t\t\tthis.mem.setInt32(sp + 16, id, true);\n\t\t\t\t\t},\n\n\t\t\t\t\t// func clearTimeoutEvent(id int32)\n\t\t\t\t\t\"runtime.clearTimeoutEvent\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\tconst id = this.mem.getInt32(sp + 8, true);\n\t\t\t\t\t\tclearTimeout(this._scheduledTimeouts.get(id));\n\t\t\t\t\t\tthis._scheduledTimeouts.delete(id);\n\t\t\t\t\t},\n\n\t\t\t\t\t// func getRandomData(r []byte)\n\t\t\t\t\t\"runtime.getRandomData\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\tcrypto.getRandomValues(loadSlice(sp + 8));\n\t\t\t\t\t},\n\n\t\t\t\t\t// func finalizeRef(v ref)\n\t\t\t\t\t\"syscall/js.finalizeRef\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\tconst id = this.mem.getUint32(sp + 8, true);\n\t\t\t\t\t\tthis._goRefCounts[id]--;\n\t\t\t\t\t\tif (this._goRefCounts[id] === 0) {\n\t\t\t\t\t\t\tconst v = this._values[id];\n\t\t\t\t\t\t\tthis._values[id] = null;\n\t\t\t\t\t\t\tthis._ids.delete(v);\n\t\t\t\t\t\t\tthis._idPool.push(id);\n\t\t\t\t\t\t}\n\t\t\t\t\t},\n\n\t\t\t\t\t// func stringVal(value string) ref\n\t\t\t\t\t\"syscall/js.stringVal\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\tstoreValue(sp + 24, loadString(sp + 8));\n\t\t\t\t\t},\n\n\t\t\t\t\t// func valueGet(v ref, p string) ref\n\t\t\t\t\t\"syscall/js.valueGet\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\tconst result = Reflect.get(loadValue(sp + 8), loadString(sp + 16));\n\t\t\t\t\t\tsp = this._inst.exports.getsp() >>> 0; // see comment above\n\t\t\t\t\t\tstoreValue(sp + 32, result);\n\t\t\t\t\t},\n\n\t\t\t\t\t// func valueSet(v ref, p string, x ref)\n\t\t\t\t\t\"syscall/js.valueSet\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\tReflect.set(loadValue(sp + 8), loadString(sp + 16), loadValue(sp + 32));\n\t\t\t\t\t},\n\n\t\t\t\t\t// func valueDelete(v ref, p string)\n\t\t\t\t\t\"syscall/js.valueDelete\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\tReflect.deleteProperty(loadValue(sp + 8), loadString(sp + 16));\n\t\t\t\t\t},\n\n\t\t\t\t\t// func valueIndex(v ref, i int) ref\n\t\t\t\t\t\"syscall/js.valueIndex\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\tstoreValue(sp + 24, Reflect.get(loadValue(sp + 8), getInt64(sp + 16)));\n\t\t\t\t\t},\n\n\t\t\t\t\t// valueSetIndex(v ref, i int, x ref)\n\t\t\t\t\t\"syscall/js.valueSetIndex\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\tReflect.set(loadValue(sp + 8), getInt64(sp + 16), loadValue(sp + 24));\n\t\t\t\t\t},\n\n\t\t\t\t\t// func valueCall(v ref, m string, args []ref) (ref, bool)\n\t\t\t\t\t\"syscall/js.valueCall\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\ttry {\n\t\t\t\t\t\t\tconst v = loadValue(sp + 8);\n\t\t\t\t\t\t\tconst m = Reflect.get(v, loadString(sp + 16));\n\t\t\t\t\t\t\tconst args = loadSliceOfValues(sp + 32);\n\t\t\t\t\t\t\tconst result = Reflect.apply(m, v, args);\n\t\t\t\t\t\t\tsp = this._inst.exports.getsp() >>> 0; // see comment above\n\t\t\t\t\t\t\tstoreValue(sp + 56, result);\n\t\t\t\t\t\t\tthis.mem.setUint8(sp + 64, 1);\n\t\t\t\t\t\t} catch (err) {\n\t\t\t\t\t\t\tsp = this._inst.exports.getsp() >>> 0; // see comment above\n\t\t\t\t\t\t\tstoreValue(sp + 56, err);\n\t\t\t\t\t\t\tthis.mem.setUint8(sp + 64, 0);\n\t\t\t\t\t\t}\n\t\t\t\t\t},\n\n\t\t\t\t\t// func valueInvoke(v ref, args []ref) (ref, bool)\n\t\t\t\t\t\"syscall/js.valueInvoke\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\ttry {\n\t\t\t\t\t\t\tconst v = loadValue(sp + 8);\n\t\t\t\t\t\t\tconst args = loadSliceOfValues(sp + 16);\n\t\t\t\t\t\t\tconst result = Reflect.apply(v, undefined, args);\n\t\t\t\t\t\t\tsp = this._inst.exports.getsp() >>> 0; // see comment above\n\t\t\t\t\t\t\tstoreValue(sp + 40, result);\n\t\t\t\t\t\t\tthis.mem.setUint8(sp + 48, 1);\n\t\t\t\t\t\t} catch (err) {\n\t\t\t\t\t\t\tsp = this._inst.exports.getsp() >>> 0; // see comment above\n\t\t\t\t\t\t\tstoreValue(sp + 40, err);\n\t\t\t\t\t\t\tthis.mem.setUint8(sp + 48, 0);\n\t\t\t\t\t\t}\n\t\t\t\t\t},\n\n\t\t\t\t\t// func valueNew(v ref, args []ref) (ref, bool)\n\t\t\t\t\t\"syscall/js.valueNew\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\ttry {\n\t\t\t\t\t\t\tconst v = loadValue(sp + 8);\n\t\t\t\t\t\t\tconst args = loadSliceOfValues(sp + 16);\n\t\t\t\t\t\t\tconst result = Reflect.construct(v, args);\n\t\t\t\t\t\t\tsp = this._inst.exports.getsp() >>> 0; // see comment above\n\t\t\t\t\t\t\tstoreValue(sp + 40, result);\n\t\t\t\t\t\t\tthis.mem.setUint8(sp + 48, 1);\n\t\t\t\t\t\t} catch (err) {\n\t\t\t\t\t\t\tsp = this._inst.exports.getsp() >>> 0; // see comment above\n\t\t\t\t\t\t\tstoreValue(sp + 40, err);\n\t\t\t\t\t\t\tthis.mem.setUint8(sp + 48, 0);\n\t\t\t\t\t\t}\n\t\t\t\t\t},\n\n\t\t\t\t\t// func valueLength(v ref) int\n\t\t\t\t\t\"syscall/js.valueLength\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\tsetInt64(sp + 16, parseInt(loadValue(sp + 8).length));\n\t\t\t\t\t},\n\n\t\t\t\t\t// valuePrepareString(v ref) (ref, int)\n\t\t\t\t\t\"syscall/js.valuePrepareString\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\tconst str = encoder.encode(String(loadValue(sp + 8)));\n\t\t\t\t\t\tstoreValue(sp + 16, str);\n\t\t\t\t\t\tsetInt64(sp + 24, str.length);\n\t\t\t\t\t},\n\n\t\t\t\t\t// valueLoadString(v ref, b []byte)\n\t\t\t\t\t\"syscall/js.valueLoadString\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\tconst str = loadValue(sp + 8);\n\t\t\t\t\t\tloadSlice(sp + 16).set(str);\n\t\t\t\t\t},\n\n\t\t\t\t\t// func valueInstanceOf(v ref, t ref) bool\n\t\t\t\t\t\"syscall/js.valueInstanceOf\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\tthis.mem.setUint8(sp + 24, (loadValue(sp + 8) instanceof loadValue(sp + 16)) ? 1 : 0);\n\t\t\t\t\t},\n\n\t\t\t\t\t// func copyBytesToGo(dst []byte, src ref) (int, bool)\n\t\t\t\t\t\"syscall/js.copyBytesToGo\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\tconst dst = loadSlice(sp + 8);\n\t\t\t\t\t\tconst src = loadValue(sp + 32);\n\t\t\t\t\t\tif (!(src instanceof Uint8Array || src instanceof Uint8ClampedArray)) {\n\t\t\t\t\t\t\tthis.mem.setUint8(sp + 48, 0);\n\t\t\t\t\t\t\treturn;\n\t\t\t\t\t\t}\n\t\t\t\t\t\tconst toCopy = src.subarray(0, dst.length);\n\t\t\t\t\t\tdst.set(toCopy);\n\t\t\t\t\t\tsetInt64(sp + 40, toCopy.length);\n\t\t\t\t\t\tthis.mem.setUint8(sp + 48, 1);\n\t\t\t\t\t},\n\n\t\t\t\t\t// func copyBytesToJS(dst ref, src []byte) (int, bool)\n\t\t\t\t\t\"syscall/js.copyBytesToJS\": (sp) => {\n\t\t\t\t\t\tsp >>>= 0;\n\t\t\t\t\t\tconst dst = loadValue(sp + 8);\n\t\t\t\t\t\tconst src = loadSlice(sp + 16);\n\t\t\t\t\t\tif (!(dst instanceof Uint8Array || dst instanceof Uint8ClampedArray)) {\n\t\t\t\t\t\t\tthis.mem.setUint8(sp + 48, 0);\n\t\t\t\t\t\t\treturn;\n\t\t\t\t\t\t}\n\t\t\t\t\t\tconst toCopy = src.subarray(0, dst.length);\n\t\t\t\t\t\tdst.set(toCopy);\n\t\t\t\t\t\tsetInt64(sp + 40, toCopy.length);\n\t\t\t\t\t\tthis.mem.setUint8(sp + 48, 1);\n\t\t\t\t\t},\n\n\t\t\t\t\t\"debug\": (value) => {\n\t\t\t\t\t\tconsole.log(value);\n\t\t\t\t\t},\n\t\t\t\t}\n\t\t\t};\n\t\t}\n\n\t\tasync run(instance) {\n\t\t\tif (!(instance instanceof WebAssembly.Instance)) {\n\t\t\t\tthrow new Error(\"Go.run: WebAssembly.Instance expected\");\n\t\t\t}\n\t\t\tthis._inst = instance;\n\t\t\tthis.mem = new DataView(this._inst.exports.mem.buffer);\n\t\t\tthis._values = [ // JS values that Go currently has references to, indexed by reference id\n\t\t\t\tNaN,\n\t\t\t\t0,\n\t\t\t\tnull,\n\t\t\t\ttrue,\n\t\t\t\tfalse,\n\t\t\t\tglobalThis,\n\t\t\t\tthis,\n\t\t\t];\n\t\t\tthis._goRefCounts = new Array(this._values.length).fill(Infinity); // number of references that Go has to a JS value, indexed by reference id\n\t\t\tthis._ids = new Map([ // mapping from JS values to reference ids\n\t\t\t\t[0, 1],\n\t\t\t\t[null, 2],\n\t\t\t\t[true, 3],\n\t\t\t\t[false, 4],\n\t\t\t\t[globalThis, 5],\n\t\t\t\t[this, 6],\n\t\t\t]);\n\t\t\tthis._idPool = [];   // unused ids that have been garbage collected\n\t\t\tthis.exited = false; // whether the Go program has exited\n\n\t\t\t// Pass command line arguments and environment variables to WebAssembly by writing them to the linear memory.\n\t\t\tlet offset = 4096;\n\n\t\t\tconst strPtr = (str) => {\n\t\t\t\tconst ptr = offset;\n\t\t\t\tconst bytes = encoder.encode(str + \"\\0\");\n\t\t\t\tnew Uint8Array(this.mem.buffer, offset, bytes.length).set(bytes);\n\t\t\t\toffset += bytes.length;\n\t\t\t\tif (offset % 8 !== 0) {\n\t\t\t\t\toffset += 8 - (offset % 8);\n\t\t\t\t}\n\t\t\t\treturn ptr;\n\t\t\t};\n\n\t\t\tconst argc = this.argv.length;\n\n\t\t\tconst argvPtrs = [];\n\t\t\tthis.argv.forEach((arg) => {\n\t\t\t\targvPtrs.push(strPtr(arg));\n\t\t\t});\n\t\t\targvPtrs.push(0);\n\n\t\t\tconst keys = Object.keys(this.env).sort();\n\t\t\tkeys.forEach((key) => {\n\t\t\t\targvPtrs.push(strPtr(`${key}=${this.env[key]}`));\n\t\t\t});\n\t\t\targvPtrs.push(0);\n\n\t\t\tconst argv = offset;\n\t\t\targvPtrs.forEach((ptr) => {\n\t\t\t\tthis.mem.setUint32(offset, ptr, true);\n\t\t\t\tthis.mem.setUint32(offset + 4, 0, true);\n\t\t\t\toffset += 8;\n\t\t\t});\n\n\t\t\t// The linker guarantees global data starts from at least wasmMinDataAddr.\n\t\t\t// Keep in sync with cmd/link/internal/ld/data.go:wasmMinDataAddr.\n\t\t\tconst wasmMinDataAddr = 4096 + 8192;\n\t\t\tif (offset >= wasmMinDataAddr) {\n\t\t\t\tthrow new Error(\"total length of command line and environment variables exceeds limit\");\n\t\t\t}\n\n\t\t\tthis._inst.exports.run(argc, argv);\n\t\t\tif (this.exited) {\n\t\t\t\tthis._resolveExitPromise();\n\t\t\t}\n\t\t\tawait this._exitPromise;\n\t\t}\n\n\t\t_resume() {\n\t\t\tif (this.exited) {\n\t\t\t\tthrow new Error(\"Go program has already exited\");\n\t\t\t}\n\t\t\tthis._inst.exports.resume();\n\t\t\tif (this.exited) {\n\t\t\t\tthis._resolveExitPromise();\n\t\t\t}\n\t\t}\n\n\t\t_makeFuncWrapper(id) {\n\t\t\tconst go = this;\n\t\t\treturn function () {\n\t\t\t\tconst event = { id: id, this: this, args: arguments };\n\t\t\t\tgo._pendingEvent = event;\n\t\t\t\tgo._resume();\n\t\t\t\treturn event.result;\n\t\t\t};\n\t\t}\n\t}\n})();\n\n\n//# sourceURL=webpack://xxdk-wasm/./node_modules/golang-wasm/src/wasm_exec.js?");

/***/ }),

/***/ "./src/index.js":
/*!**********************!*\
  !*** ./src/index.js ***!
  \**********************/
/***/ ((module, __webpack_exports__, __webpack_require__) => {

eval("__webpack_require__.a(module, async (__webpack_handle_async_dependencies__, __webpack_async_result__) => { try {\n__webpack_require__.r(__webpack_exports__);\n/* harmony import */ var _api_main_go__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! ./api/main.go */ \"./src/api/main.go\");\n////////////////////////////////////////////////////////////////////////////////\n// Copyright © 2022 xx foundation                                             //\n//                                                                            //\n// Use of this source code is governed by a license that can be found in the  //\n// LICENSE file.                                                              //\n////////////////////////////////////////////////////////////////////////////////\n\n// import _ from 'lodash';\n// import { divide } from \"./api/main.go\";\n\n\n\nconst { divide } = _api_main_go__WEBPACK_IMPORTED_MODULE_0__[\"default\"];\n\n// function component() {\n//     const element = document.createElement('div');\n//\n//     // Lodash, now imported by this script\n//     element.innerHTML = _.join(['Hello', 'webpack'], ' ');\n//\n//     return element;\n// }\n//\n// document.body.appendChild(component());\n\nconst result = await divide(6, 2);\nconsole.log(result); // 3\n\n__webpack_async_result__();\n} catch(e) { __webpack_async_result__(e); } }, 1);\n\n//# sourceURL=webpack://xxdk-wasm/./src/index.js?");

/***/ })

/******/ 	});
/************************************************************************/
/******/ 	// The module cache
/******/ 	var __webpack_module_cache__ = {};
/******/ 	
/******/ 	// The require function
/******/ 	function __webpack_require__(moduleId) {
/******/ 		// Check if module is in cache
/******/ 		var cachedModule = __webpack_module_cache__[moduleId];
/******/ 		if (cachedModule !== undefined) {
/******/ 			return cachedModule.exports;
/******/ 		}
/******/ 		// Create a new module (and put it into the cache)
/******/ 		var module = __webpack_module_cache__[moduleId] = {
/******/ 			// no module.id needed
/******/ 			// no module.loaded needed
/******/ 			exports: {}
/******/ 		};
/******/ 	
/******/ 		// Execute the module function
/******/ 		__webpack_modules__[moduleId](module, module.exports, __webpack_require__);
/******/ 	
/******/ 		// Return the exports of the module
/******/ 		return module.exports;
/******/ 	}
/******/ 	
/************************************************************************/
/******/ 	/* webpack/runtime/async module */
/******/ 	(() => {
/******/ 		var webpackQueues = typeof Symbol === "function" ? Symbol("webpack queues") : "__webpack_queues__";
/******/ 		var webpackExports = typeof Symbol === "function" ? Symbol("webpack exports") : "__webpack_exports__";
/******/ 		var webpackError = typeof Symbol === "function" ? Symbol("webpack error") : "__webpack_error__";
/******/ 		var resolveQueue = (queue) => {
/******/ 			if(queue && !queue.d) {
/******/ 				queue.d = 1;
/******/ 				queue.forEach((fn) => (fn.r--));
/******/ 				queue.forEach((fn) => (fn.r-- ? fn.r++ : fn()));
/******/ 			}
/******/ 		}
/******/ 		var wrapDeps = (deps) => (deps.map((dep) => {
/******/ 			if(dep !== null && typeof dep === "object") {
/******/ 				if(dep[webpackQueues]) return dep;
/******/ 				if(dep.then) {
/******/ 					var queue = [];
/******/ 					queue.d = 0;
/******/ 					dep.then((r) => {
/******/ 						obj[webpackExports] = r;
/******/ 						resolveQueue(queue);
/******/ 					}, (e) => {
/******/ 						obj[webpackError] = e;
/******/ 						resolveQueue(queue);
/******/ 					});
/******/ 					var obj = {};
/******/ 					obj[webpackQueues] = (fn) => (fn(queue));
/******/ 					return obj;
/******/ 				}
/******/ 			}
/******/ 			var ret = {};
/******/ 			ret[webpackQueues] = x => {};
/******/ 			ret[webpackExports] = dep;
/******/ 			return ret;
/******/ 		}));
/******/ 		__webpack_require__.a = (module, body, hasAwait) => {
/******/ 			var queue;
/******/ 			hasAwait && ((queue = []).d = 1);
/******/ 			var depQueues = new Set();
/******/ 			var exports = module.exports;
/******/ 			var currentDeps;
/******/ 			var outerResolve;
/******/ 			var reject;
/******/ 			var promise = new Promise((resolve, rej) => {
/******/ 				reject = rej;
/******/ 				outerResolve = resolve;
/******/ 			});
/******/ 			promise[webpackExports] = exports;
/******/ 			promise[webpackQueues] = (fn) => (queue && fn(queue), depQueues.forEach(fn), promise["catch"](x => {}));
/******/ 			module.exports = promise;
/******/ 			body((deps) => {
/******/ 				currentDeps = wrapDeps(deps);
/******/ 				var fn;
/******/ 				var getResult = () => (currentDeps.map((d) => {
/******/ 					if(d[webpackError]) throw d[webpackError];
/******/ 					return d[webpackExports];
/******/ 				}))
/******/ 				var promise = new Promise((resolve) => {
/******/ 					fn = () => (resolve(getResult));
/******/ 					fn.r = 0;
/******/ 					var fnQueue = (q) => (q !== queue && !depQueues.has(q) && (depQueues.add(q), q && !q.d && (fn.r++, q.push(fn))));
/******/ 					currentDeps.map((dep) => (dep[webpackQueues](fnQueue)));
/******/ 				});
/******/ 				return fn.r ? promise : getResult();
/******/ 			}, (err) => ((err ? reject(promise[webpackError] = err) : outerResolve(exports)), resolveQueue(queue)));
/******/ 			queue && (queue.d = 0);
/******/ 		};
/******/ 	})();
/******/ 	
/******/ 	/* webpack/runtime/define property getters */
/******/ 	(() => {
/******/ 		// define getter functions for harmony exports
/******/ 		__webpack_require__.d = (exports, definition) => {
/******/ 			for(var key in definition) {
/******/ 				if(__webpack_require__.o(definition, key) && !__webpack_require__.o(exports, key)) {
/******/ 					Object.defineProperty(exports, key, { enumerable: true, get: definition[key] });
/******/ 				}
/******/ 			}
/******/ 		};
/******/ 	})();
/******/ 	
/******/ 	/* webpack/runtime/global */
/******/ 	(() => {
/******/ 		__webpack_require__.g = (function() {
/******/ 			if (typeof globalThis === 'object') return globalThis;
/******/ 			try {
/******/ 				return this || new Function('return this')();
/******/ 			} catch (e) {
/******/ 				if (typeof window === 'object') return window;
/******/ 			}
/******/ 		})();
/******/ 	})();
/******/ 	
/******/ 	/* webpack/runtime/hasOwnProperty shorthand */
/******/ 	(() => {
/******/ 		__webpack_require__.o = (obj, prop) => (Object.prototype.hasOwnProperty.call(obj, prop))
/******/ 	})();
/******/ 	
/******/ 	/* webpack/runtime/make namespace object */
/******/ 	(() => {
/******/ 		// define __esModule on exports
/******/ 		__webpack_require__.r = (exports) => {
/******/ 			if(typeof Symbol !== 'undefined' && Symbol.toStringTag) {
/******/ 				Object.defineProperty(exports, Symbol.toStringTag, { value: 'Module' });
/******/ 			}
/******/ 			Object.defineProperty(exports, '__esModule', { value: true });
/******/ 		};
/******/ 	})();
/******/ 	
/************************************************************************/
/******/ 	
/******/ 	// startup
/******/ 	// Load entry module and return exports
/******/ 	// This entry module can't be inlined because the eval devtool is used.
/******/ 	var __webpack_exports__ = __webpack_require__("./src/index.js");
/******/ 	
/******/ })()
;