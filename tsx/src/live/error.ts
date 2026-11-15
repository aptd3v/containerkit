export const errorClasses = [
	"notFound",
	"alreadyExists",
	"conflict",
	"unauthorized",
	"permissionDenied",
	"invalidArgument",
	"unavailable",
	"loginFailed",
	"loginRequired",
	"parsingFailed",
	"canceled",
	"internal",
	"unknown",
	"failedPrecondition",
	"notImplemented",
	"resourceExhausted",
	"aborted",
	"outOfRange",
	"dataLoss",
	"unsupported",
	"unsupportedFlag",
	"incompatible",
	"disabled",
	"invalidProject",
	"wrongContextType",
	"notModified",
] as const;

export type ErrorClass = (typeof errorClasses)[number];

export type ErrorFrame = {
	type: "error";
	class: string;
	message: string;
	status?: number;
	hint?: string;
	receiver?: string;
	path?: string;
	exitCode?: number;
};

export class ContainerKitError extends Error {
	readonly class: string;
	readonly status: number;
	readonly hint: string;
	readonly receiver: string;
	readonly path: string;
	readonly exitCode?: number;

	constructor(frame: ErrorFrame) {
		super(frame.message);
		this.name = "ContainerKitError";
		this.class = frame.class;
		this.status = frame.status ?? 0;
		this.hint = frame.hint ?? "";
		this.receiver = frame.receiver ?? "";
		this.path = frame.path ?? "";
		this.exitCode = frame.exitCode;
	}
}

function matches(err: unknown, cls: ErrorClass): err is ContainerKitError {
	return err instanceof ContainerKitError && err.class === cls;
}

export function isCKNotFoundError(err: unknown): err is ContainerKitError {
	return matches(err, "notFound");
}
export function isCKAlreadyExistsError(err: unknown): err is ContainerKitError {
	return matches(err, "alreadyExists");
}
export function isCKConflictError(err: unknown): err is ContainerKitError {
	return matches(err, "conflict");
}
export function isCKUnauthorizedError(err: unknown): err is ContainerKitError {
	return matches(err, "unauthorized");
}
export function isCKPermissionDeniedError(err: unknown): err is ContainerKitError {
	return matches(err, "permissionDenied");
}
export function isCKInvalidArgumentError(err: unknown): err is ContainerKitError {
	return matches(err, "invalidArgument");
}
export function isCKUnavailableError(err: unknown): err is ContainerKitError {
	return matches(err, "unavailable");
}
export function isCKLoginFailedError(err: unknown): err is ContainerKitError {
	return matches(err, "loginFailed");
}
export function isCKLoginRequiredError(err: unknown): err is ContainerKitError {
	return matches(err, "loginRequired");
}
export function isCKParsingFailedError(err: unknown): err is ContainerKitError {
	return matches(err, "parsingFailed");
}
export function isCKCanceledError(err: unknown): err is ContainerKitError {
	return matches(err, "canceled");
}
export function isCKInternalError(err: unknown): err is ContainerKitError {
	return matches(err, "internal");
}
export function isCKUnknownError(err: unknown): err is ContainerKitError {
	return matches(err, "unknown");
}
export function isCKFailedPreconditionError(err: unknown): err is ContainerKitError {
	return matches(err, "failedPrecondition");
}
export function isCKNotImplementedError(err: unknown): err is ContainerKitError {
	return matches(err, "notImplemented");
}
export function isCKResourceExhaustedError(err: unknown): err is ContainerKitError {
	return matches(err, "resourceExhausted");
}
export function isCKAbortedError(err: unknown): err is ContainerKitError {
	return matches(err, "aborted");
}
export function isCKOutOfRangeError(err: unknown): err is ContainerKitError {
	return matches(err, "outOfRange");
}
export function isCKDataLossError(err: unknown): err is ContainerKitError {
	return matches(err, "dataLoss");
}
export function isCKUnsupportedError(err: unknown): err is ContainerKitError {
	return matches(err, "unsupported");
}
export function isCKUnsupportedFlagError(err: unknown): err is ContainerKitError {
	return matches(err, "unsupportedFlag");
}
export function isCKIncompatibleError(err: unknown): err is ContainerKitError {
	return matches(err, "incompatible");
}
export function isCKDisabledError(err: unknown): err is ContainerKitError {
	return matches(err, "disabled");
}
export function isCKInvalidProjectError(err: unknown): err is ContainerKitError {
	return matches(err, "invalidProject");
}
export function isCKWrongContextTypeError(err: unknown): err is ContainerKitError {
	return matches(err, "wrongContextType");
}
export function isCKNotModifiedError(err: unknown): err is ContainerKitError {
	return matches(err, "notModified");
}

export function errorFromFrame(frame: ErrorFrame): ContainerKitError {
	return new ContainerKitError(frame);
}
