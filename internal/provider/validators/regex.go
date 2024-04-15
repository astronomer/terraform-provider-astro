package validators

const KubernetesResourceString = `^(\+)?((\d+(\.\d*)?)|(\.\d+))(([KMGTPE]i)|[mkMGTPE]|([eE](\+)?((\d+(\.\d*)?)|(\.\d+))))?$`
const EmailString = `^.+@.+\..+$` // May let some invalid emails through but should be enough for most cases
