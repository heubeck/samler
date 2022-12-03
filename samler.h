#ifndef SML_CGO_H_
#define SML_CGO_H_

struct DeviceConfig {
    const char *name;
    int baudRate;
    const char *mode;
};

struct SmlValue{
    const char *value;
    const char *unit;
    const char *prefix;
    const char *ident;
    const char *suffix;
};

struct SmlData{
    struct SmlValue value;
};

typedef void (*SmlEvent)(struct SmlData message);

typedef struct {
    SmlEvent event;
} Callbacks;

int listen_to_device(struct DeviceConfig config, Callbacks callbacks);

extern void onSmlMessage(struct SmlData);
void propagateEvent(struct SmlData message);

#endif
