export type PermissionLevel = 'ask' | 'allow' | 'deny';

export class PermissionManager {
	private allowedTools: Set<string>;
	private deniedTools: Set<string>;

	constructor(allowedTools: string[] = []) {
		this.allowedTools = new Set(allowedTools);
		this.deniedTools = new Set();
	}

	checkPermission(toolName: string): PermissionLevel {
		if (this.deniedTools.has(toolName)) return 'deny';

		// Wildcard support e.g. "git-*"
		for (const allowed of this.allowedTools) {
			if (allowed === '*' || allowed === toolName) return 'allow';
			if (allowed.endsWith('*') && toolName.startsWith(allowed.slice(0, -1)))
				return 'allow';
		}

		return 'ask';
	}

	grantPermission(toolName: string) {
		this.allowedTools.add(toolName);
	}

	denyPermission(toolName: string) {
		this.deniedTools.add(toolName);
	}
}
