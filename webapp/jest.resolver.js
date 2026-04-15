// Custom Jest resolver to handle the node: protocol (e.g. node:stream, node:buffer).
// Jest 26 does not support this natively; strip the prefix and resolve the built-in normally.
module.exports = (request, options) => {
    return options.defaultResolver(
        request.startsWith('node:') ? request.slice(5) : request,
        options,
    );
};
