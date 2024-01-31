import './wasm_exec_tinygo.js';
import mod from './wasm.wasm';

const go = new Go();
const instance = await WebAssembly.instantiate(mod, go.importObject);

export default {
	/**
	 * @param {Request} request
	 */
	async fetch(request) {
		const params = new URL(request.url).searchParams;
		const items = params.get('items');

		const p = instance.exports.getPtr(items.length);
		const l = instance.exports.getLength();

		const slice = new Uint8Array(instance.exports.memory.buffer, p, l + 1);

		const buffer = new TextEncoder().encode(items);
		slice.set(buffer);
		slice[buffer.length] = 0;

		const ptr = instance.exports.buildGif();
		const size = instance.exports.getLength();

		const mem = new Int8Array(instance.exports.memory.buffer);
		const view = mem.subarray(ptr, ptr + size);

		const blob = new Blob([view], {
			type: 'image/gif',
		});

		return new Response(blob, {
			headers: {
				'Content-Type': 'image/gif',
			},
		});
	},
};
