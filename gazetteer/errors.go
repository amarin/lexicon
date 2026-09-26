package gazetteer

import "errors"

// ErrUnknownSource is returned by RefreshSource for a name not in Config.Sources.
var ErrUnknownSource = errors.New("gazetteer: unknown source")
