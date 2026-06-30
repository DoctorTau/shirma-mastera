export function parseHpNumber(text?: string): number | undefined {
	if (!text) return undefined;
	const m = text.match(/\d+/);
	return m ? Number(m[0]) : undefined;
}
