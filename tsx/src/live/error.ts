const errorClasses = [
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

/**
 * Stable class of a session failure. Switch on it.
 * A class string the session does not know is reported as `"unknown"`.
 */
export type ErrorClass = (typeof errorClasses)[number];

const knownClasses: ReadonlySet<string> = new Set(errorClasses);

/** Maps a session class string onto ErrorClass. An unrecognized value becomes `"unknown"`. */
export function errorClass(value: string): ErrorClass {
	return knownClasses.has(value) ? (value as ErrorClass) : "unknown";
}

/** One error frame from the session. `message` has secrets removed. */
export type ErrorFrame = {
	type: "error";
	/** Which failure this is. */
	class: ErrorClass;
	/** What failed, with secrets removed. */
	message: string;
	/** Numeric code from the daemon, when it sent one. */
	status?: number;
	/** Short suggestion for the next step, when there is one. */
	hint?: string;
	/** Area that failed, such as a container or a network. */
	receiver?: string;
	/** Setting or file involved, when there is one. */
	path?: string;
	/** Container exit status, when the failure is a container exiting. */
	exitCode?: number;
};

/** A session error before its class string is checked. An unknown class becomes `"unknown"`. */
export type ErrorReport = Omit<ErrorFrame, "class"> & { class: string };

/**
 * Failure from a session command.
 *
 * @remarks
 * Session commands reject with this error. They do not resolve it.
 * `class` is the closed set in {@link ErrorClass}. A class string the session does not know is `"unknown"`.
 * `message` has secrets removed.
 *
 * @template C - One {@link ErrorClass} value, so a check can narrow `class`
 *
 * @example
 * ```ts
 * try {
 *   await session.inspect("missing");
 * } catch (err) {
 *   if (isContainerKitError(err, "notFound")) {
 *     err.class; // "notFound"
 *   }
 * }
 * ```
 *
 * @see {@link isContainerKitError}
 */
export class ContainerKitError<C extends ErrorClass = ErrorClass> extends Error {
	/** Which failure this is. */
	readonly class: C;
	/** Numeric code from the daemon. `0` when the daemon sent none. */
	readonly status: number;
	/** Short suggestion for the next step. Empty when there is none. */
	readonly hint: string;
	/** Area that failed, such as a container or a network. Empty when there is none. */
	readonly receiver: string;
	/** Setting or file involved. Empty when there is none. */
	readonly path: string;
	/** Container exit status, when the failure is a container exiting. */
	readonly exitCode?: number;

	constructor(frame: ErrorFrame & { class: C }) {
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

/**
 * True when `err` is a session failure.
 *
 * @param err - A thrown value
 * @param cls - When set, the check is true only for this {@link ErrorClass}, and `err.class` narrows to it
 * @returns Whether `err` is a {@link ContainerKitError} of that class
 *
 * @example
 * ```ts
 * if (isContainerKitError(err, "unavailable")) {
 *   return;
 * }
 * ```
 */
export function isContainerKitError(err: unknown): err is ContainerKitError;
export function isContainerKitError<C extends ErrorClass>(err: unknown, cls: C): err is ContainerKitError<C>;
export function isContainerKitError(err: unknown, cls?: ErrorClass): err is ContainerKitError {
	if (!(err instanceof ContainerKitError)) {
		return false;
	}
	return cls === undefined || err.class === cls;
}

function isClass<C extends ErrorClass>(cls: C): (err: unknown) => err is ContainerKitError<C> {
	return (err): err is ContainerKitError<C> => isContainerKitError(err, cls);
}

/** The image, container, network, volume, or other object is not there. */
export const isCKNotFoundError = isClass("notFound");
/** It is already there. */
export const isCKAlreadyExistsError = isClass("alreadyExists");
/** The request clashes with the current state. */
export const isCKConflictError = isClass("conflict");
/** The daemon rejected the credentials. */
export const isCKUnauthorizedError = isClass("unauthorized");
/** The credentials are known and are not allowed to do this. */
export const isCKPermissionDeniedError = isClass("permissionDenied");
/** A value in the command is not acceptable. */
export const isCKInvalidArgumentError = isClass("invalidArgument");
/** Docker cannot be reached. The session closes after `onError`. */
export const isCKUnavailableError = isClass("unavailable");
/** The login was attempted and did not succeed. */
export const isCKLoginFailedError = isClass("loginFailed");
/** A registry needs a login before the pull or push. */
export const isCKLoginRequiredError = isClass("loginRequired");
/** The YAML, Dockerfile, or other text could not be parsed. */
export const isCKParsingFailedError = isClass("parsingFailed");
/** The call was canceled. */
export const isCKCanceledError = isClass("canceled");
/** The failure is inside the daemon or the session. */
export const isCKInternalError = isClass("internal");
/** The failure did not fit another check. */
export const isCKUnknownError = isClass("unknown");
/** This call is legal only after something else has happened. */
export const isCKFailedPreconditionError = isClass("failedPrecondition");
/** The operation is not implemented. */
export const isCKNotImplementedError = isClass("notImplemented");
/** The daemon is out of space, memory, or a similar limit. */
export const isCKResourceExhaustedError = isClass("resourceExhausted");
/** The call was aborted. */
export const isCKAbortedError = isClass("aborted");
/** A number or index is outside the range the call accepts. */
export const isCKOutOfRangeError = isClass("outOfRange");
/** The result was lost. */
export const isCKDataLossError = isClass("dataLoss");
/** This daemon or this call does not support the operation. */
export const isCKUnsupportedError = isClass("unsupported");
/** A flag on the call is not supported here. */
export const isCKUnsupportedFlagError = isClass("unsupportedFlag");
/** The daemon and the request do not agree. */
export const isCKIncompatibleError = isClass("incompatible");
/** The feature is turned off. */
export const isCKDisabledError = isClass("disabled");
/** The project description itself is not valid. */
export const isCKInvalidProjectError = isClass("invalidProject");
/** The command was aimed at the wrong kind of object. */
export const isCKWrongContextTypeError = isClass("wrongContextType");
/** Nothing changed. */
export const isCKNotModifiedError = isClass("notModified");

/** Builds a {@link ContainerKitError} from a session frame. An unknown class becomes `"unknown"`. */
export function errorFromFrame(frame: ErrorReport): ContainerKitError {
	return new ContainerKitError({ ...frame, class: errorClass(frame.class) });
}
