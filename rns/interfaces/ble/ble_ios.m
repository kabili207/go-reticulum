#import <Foundation/Foundation.h>
#import <CoreBluetooth/CoreBluetooth.h>

static CBUUID* NUSServiceUUID(void) {
  return [CBUUID UUIDWithString:@"6E400001-B5A3-F393-E0A9-E50E24DCCA9E"];
}
static CBUUID* NUSRXUUID(void) {
  return [CBUUID UUIDWithString:@"6E400002-B5A3-F393-E0A9-E50E24DCCA9E"];
}
static CBUUID* NUSTXUUID(void) {
  return [CBUUID UUIDWithString:@"6E400003-B5A3-F393-E0A9-E50E24DCCA9E"];
}

@interface ReticulumBLE : NSObject<CBCentralManagerDelegate, CBPeripheralDelegate, CBPeripheralManagerDelegate>
@property(nonatomic, strong) CBCentralManager* central;
@property(nonatomic, strong) dispatch_queue_t queue;
@property(nonatomic, strong) dispatch_semaphore_t readySem;
@property(nonatomic, strong) dispatch_semaphore_t openSem;

@property(nonatomic, strong) CBPeripheralManager* peripheralMgr;
@property(nonatomic, strong) dispatch_semaphore_t peripheralReadySem;
@property(nonatomic) BOOL advertising;
@property(nonatomic, strong) CBMutableCharacteristic* periphRX;
@property(nonatomic, strong) CBMutableCharacteristic* periphTX;
@property(nonatomic, strong) NSMutableArray<CBCentral*>* subscribers;
@property(nonatomic, strong) NSMutableData* txPending;

@property(nonatomic, strong) CBPeripheral* peripheral;
@property(nonatomic, strong) CBCharacteristic* rxChar;
@property(nonatomic, strong) CBCharacteristic* txChar;

@property(nonatomic, strong) NSMutableData* rxBuf;
@property(nonatomic) BOOL connected;
@property(nonatomic) BOOL disappeared;

@property(nonatomic, copy) NSString* targetName;
@property(nonatomic, copy) NSString* targetAddr;
@property(nonatomic, copy) NSString* debugStr;
@end

@implementation ReticulumBLE

- (instancetype)initWithTargetName:(NSString*)name targetAddr:(NSString*)addr {
  self = [super init];
  if (self) {
    _targetName = name ?: @"";
    _targetAddr = addr ?: @"";
    _rxBuf = [NSMutableData data];
    _connected = NO;
    _disappeared = NO;
    _queue = dispatch_queue_create("reticulum.ble", DISPATCH_QUEUE_SERIAL);
    _readySem = dispatch_semaphore_create(0);
    _openSem = dispatch_semaphore_create(0);
    _peripheralReadySem = dispatch_semaphore_create(0);
    _advertising = NO;
    _subscribers = [NSMutableArray array];
    _txPending = [NSMutableData data];
    _central = [[CBCentralManager alloc] initWithDelegate:self queue:_queue];
    _peripheralMgr = [[CBPeripheralManager alloc] initWithDelegate:self queue:_queue];
    _debugStr = @"ble://";
  }
  return self;
}

- (BOOL)matchesPeripheral:(CBPeripheral*)p adv:(NSDictionary*)adv {
  NSString* name = adv[CBAdvertisementDataLocalNameKey];
  if (name.length == 0) {
    name = p.name;
  }

  if (_targetAddr.length > 0) {
    NSString* ident = p.identifier.UUIDString.lowercaseString;
    return [ident containsString:_targetAddr.lowercaseString];
  }

  if (_targetName.length > 0) {
    return [name isEqualToString:_targetName];
  }

  return [name hasPrefix:@"RNode "];
}

- (void)centralManagerDidUpdateState:(CBCentralManager*)central {
  dispatch_semaphore_signal(_readySem);
}

- (void)peripheralManagerDidUpdateState:(CBPeripheralManager*)peripheral {
  dispatch_semaphore_signal(_peripheralReadySem);
  if (peripheral.state != CBManagerStatePoweredOn) {
    self.advertising = NO;
    return;
  }

  // Expose the NUS service so other devices can connect to iOS and exchange bytes.
  CBUUID* svcUUID = NUSServiceUUID();
  CBUUID* rxUUID = NUSRXUUID();
  CBUUID* txUUID = NUSTXUUID();

  self.periphRX = [[CBMutableCharacteristic alloc] initWithType:rxUUID
                                                    properties:(CBCharacteristicPropertyWriteWithoutResponse | CBCharacteristicPropertyWrite)
                                                         value:nil
                                                   permissions:(CBAttributePermissionsWriteable)];

  self.periphTX = [[CBMutableCharacteristic alloc] initWithType:txUUID
                                                    properties:(CBCharacteristicPropertyNotify)
                                                         value:nil
                                                   permissions:(CBAttributePermissionsReadable)];

  CBMutableService* service = [[CBMutableService alloc] initWithType:svcUUID primary:YES];
  service.characteristics = @[ self.periphRX, self.periphTX ];
  [self.peripheralMgr removeAllServices];
  [self.peripheralMgr addService:service];
}

- (void)peripheralManager:(CBPeripheralManager*)peripheral didAddService:(CBService*)service error:(NSError*)error {
  if (error != nil) {
    self.advertising = NO;
    return;
  }
  // Advertising: include service UUID, and a best-effort local name matching the default "RNode " prefix.
  NSDictionary* adv = @{
    CBAdvertisementDataServiceUUIDsKey : @[ NUSServiceUUID() ],
    CBAdvertisementDataLocalNameKey : @"RNode iOS"
  };
  [self.peripheralMgr startAdvertising:adv];
  self.advertising = YES;
}

- (void)peripheralManagerDidStartAdvertising:(CBPeripheralManager*)peripheral error:(NSError*)error {
  if (error != nil) {
    self.advertising = NO;
  }
}

- (void)peripheralManager:(CBPeripheralManager*)peripheral central:(CBCentral*)central didSubscribeToCharacteristic:(CBCharacteristic*)characteristic {
  if (![characteristic.UUID isEqual:NUSTXUUID()]) return;
  if (central != nil) {
    [self.subscribers addObject:central];
  }
}

- (void)peripheralManager:(CBPeripheralManager*)peripheral central:(CBCentral*)central didUnsubscribeFromCharacteristic:(CBCharacteristic*)characteristic {
  if (![characteristic.UUID isEqual:NUSTXUUID()]) return;
  if (central != nil) {
    [self.subscribers removeObject:central];
  }
}

- (void)peripheralManager:(CBPeripheralManager*)peripheral didReceiveWriteRequests:(NSArray<CBATTRequest*>*)requests {
  for (CBATTRequest* req in requests) {
    if (![req.characteristic.UUID isEqual:NUSRXUUID()]) {
      continue;
    }
    NSData* v = req.value;
    if (v.length > 0) {
      @synchronized(self.rxBuf) {
        [self.rxBuf appendData:v];
      }
    }
    [peripheral respondToRequest:req withResult:CBATTErrorSuccess];
  }
}

- (void)peripheralManagerIsReadyToUpdateSubscribers:(CBPeripheralManager*)peripheral {
  // Try flushing any pending TX data when CoreBluetooth says it can accept more updates.
  if (self.txPending.length == 0 || self.periphTX == nil) return;
  [self flushPeripheralTX];
}

- (void)flushPeripheralTX {
  if (!self.advertising || self.periphTX == nil || self.subscribers.count == 0) return;

  // CoreBluetooth provides the per-central maximum update length.
  NSInteger maxLen = 0;
  for (CBCentral* c in self.subscribers) {
    if (c == nil) continue;
    NSInteger n = (NSInteger)c.maximumUpdateValueLength;
    if (n > 0 && (maxLen == 0 || n < maxLen)) {
      maxLen = n;
    }
  }
  if (maxLen <= 0) maxLen = 20;

  while (self.txPending.length > 0) {
    NSInteger n = MIN(maxLen, (NSInteger)self.txPending.length);
    NSData* chunk = [self.txPending subdataWithRange:NSMakeRange(0, (NSUInteger)n)];
    BOOL ok = [self.peripheralMgr updateValue:chunk forCharacteristic:self.periphTX onSubscribedCentrals:self.subscribers];
    if (!ok) {
      return;
    }
    [self.txPending replaceBytesInRange:NSMakeRange(0, (NSUInteger)n) withBytes:NULL length:0];
  }
}

- (void)centralManager:(CBCentralManager*)central
 didDiscoverPeripheral:(CBPeripheral*)peripheral
     advertisementData:(NSDictionary<NSString*, id>*)advertisementData
                  RSSI:(NSNumber*)RSSI {
  if (self.peripheral != nil) {
    return;
  }
  if (![self matchesPeripheral:peripheral adv:advertisementData]) {
    return;
  }
  self.peripheral = peripheral;
  self.peripheral.delegate = self;
  self.debugStr = [NSString stringWithFormat:@"ble://%@", peripheral.identifier.UUIDString];
  [self.central stopScan];
  [self.central connectPeripheral:peripheral options:nil];
}

- (void)centralManager:(CBCentralManager*)central didConnectPeripheral:(CBPeripheral*)peripheral {
  [peripheral discoverServices:@[NUSServiceUUID()]];
}

- (void)centralManager:(CBCentralManager*)central didFailToConnectPeripheral:(CBPeripheral*)peripheral error:(NSError*)error {
  self.connected = NO;
  dispatch_semaphore_signal(_openSem);
}

- (void)centralManager:(CBCentralManager*)central didDisconnectPeripheral:(CBPeripheral*)peripheral error:(NSError*)error {
  self.connected = NO;
  self.disappeared = YES;
  self.rxChar = nil;
  self.txChar = nil;
}

- (void)peripheral:(CBPeripheral*)peripheral didDiscoverServices:(NSError*)error {
  if (error != nil) {
    dispatch_semaphore_signal(_openSem);
    return;
  }
  for (CBService* svc in peripheral.services) {
    if ([svc.UUID isEqual:NUSServiceUUID()]) {
      [peripheral discoverCharacteristics:@[NUSRXUUID(), NUSTXUUID()] forService:svc];
      return;
    }
  }
  dispatch_semaphore_signal(_openSem);
}

- (void)peripheral:(CBPeripheral*)peripheral didDiscoverCharacteristicsForService:(CBService*)service error:(NSError*)error {
  if (error != nil) {
    dispatch_semaphore_signal(_openSem);
    return;
  }
  for (CBCharacteristic* c in service.characteristics) {
    if ([c.UUID isEqual:NUSRXUUID()]) {
      self.rxChar = c;
    } else if ([c.UUID isEqual:NUSTXUUID()]) {
      self.txChar = c;
    }
  }
  if (self.rxChar != nil && self.txChar != nil) {
    [peripheral setNotifyValue:YES forCharacteristic:self.txChar];
  } else {
    dispatch_semaphore_signal(_openSem);
  }
}

- (void)peripheral:(CBPeripheral*)peripheral didUpdateNotificationStateForCharacteristic:(CBCharacteristic*)characteristic error:(NSError*)error {
  if (error == nil && [characteristic.UUID isEqual:NUSTXUUID()] && characteristic.isNotifying) {
    self.connected = YES;
  }
  dispatch_semaphore_signal(_openSem);
}

- (void)peripheral:(CBPeripheral*)peripheral didUpdateValueForCharacteristic:(CBCharacteristic*)characteristic error:(NSError*)error {
  if (error != nil) {
    return;
  }
  if (![characteristic.UUID isEqual:NUSTXUUID()]) {
    return;
  }
  NSData* value = characteristic.value;
  if (value.length == 0) {
    return;
  }
  @synchronized(self.rxBuf) {
    [self.rxBuf appendData:value];
  }
}

@end

typedef void* reticulum_ble_t;

reticulum_ble_t reticulum_ble_new(const char* target_name, const char* target_addr) {
  NSString* name = target_name ? [NSString stringWithUTF8String:target_name] : @"";
  NSString* addr = target_addr ? [NSString stringWithUTF8String:target_addr] : @"";
  ReticulumBLE* obj = [[ReticulumBLE alloc] initWithTargetName:name targetAddr:addr];
  return (__bridge_retained void*)obj;
}

int reticulum_ble_open(reticulum_ble_t h, double timeout_seconds) {
  if (h == NULL) return -1;
  ReticulumBLE* obj = (__bridge ReticulumBLE*)h;

  // Wait for CBCentralManager state update.
  dispatch_time_t readyTO = dispatch_time(DISPATCH_TIME_NOW, (int64_t)(timeout_seconds * NSEC_PER_SEC));
  dispatch_semaphore_wait(obj.readySem, readyTO);
  dispatch_semaphore_wait(obj.peripheralReadySem, readyTO);

  if (obj.central.state != CBManagerStatePoweredOn && obj.peripheralMgr.state != CBManagerStatePoweredOn) {
    return -2;
  }

  obj.disappeared = NO;
  obj.connected = NO;
  obj.peripheral = nil;
  obj.rxChar = nil;
  obj.txChar = nil;

  if (obj.central.state == CBManagerStatePoweredOn) {
    [obj.central scanForPeripheralsWithServices:@[NUSServiceUUID()] options:nil];
  }

  // If a specific target is requested, block until we connect (like other platforms).
  // Otherwise, succeed once BLE is up (peripheral advertising is best-effort).
  if (obj.targetName.length > 0 || obj.targetAddr.length > 0) {
    dispatch_time_t openTO = dispatch_time(DISPATCH_TIME_NOW, (int64_t)(timeout_seconds * NSEC_PER_SEC));
    long rc = dispatch_semaphore_wait(obj.openSem, openTO);
    if (rc != 0) {
      [obj.central stopScan];
      return -3; // timeout
    }
    return obj.connected ? 0 : -4;
  }

  return 0;
}

void reticulum_ble_close(reticulum_ble_t h) {
  if (h == NULL) return;
  ReticulumBLE* obj = (__bridge ReticulumBLE*)h;
  [obj.central stopScan];
  if (obj.peripheral != nil) {
    [obj.central cancelPeripheralConnection:obj.peripheral];
  }
  if (obj.peripheralMgr != nil) {
    [obj.peripheralMgr stopAdvertising];
  }
  obj.connected = NO;
  obj.advertising = NO;
  [obj.subscribers removeAllObjects];
  @synchronized(obj.txPending) {
    [obj.txPending setLength:0];
  }
}

void reticulum_ble_free(reticulum_ble_t h) {
  if (h == NULL) return;
  // Balance __bridge_retained in reticulum_ble_new
  CFBridgingRelease(h);
}

int reticulum_ble_is_open(reticulum_ble_t h) {
  if (h == NULL) return 0;
  ReticulumBLE* obj = (__bridge ReticulumBLE*)h;
  return (obj.connected || obj.advertising) ? 1 : 0;
}

int reticulum_ble_device_disappeared(reticulum_ble_t h) {
  if (h == NULL) return 0;
  ReticulumBLE* obj = (__bridge ReticulumBLE*)h;
  return obj.disappeared ? 1 : 0;
}

int reticulum_ble_read(reticulum_ble_t h, uint8_t* out, int out_cap) {
  if (h == NULL) return -1;
  if (out == NULL || out_cap <= 0) return 0;
  ReticulumBLE* obj = (__bridge ReticulumBLE*)h;
  @synchronized(obj.rxBuf) {
    if (obj.rxBuf.length == 0) {
      return obj.connected ? 0 : -1;
    }
    int n = (int)MIN((NSUInteger)out_cap, obj.rxBuf.length);
    memcpy(out, obj.rxBuf.bytes, (size_t)n);
    [obj.rxBuf replaceBytesInRange:NSMakeRange(0, (NSUInteger)n) withBytes:NULL length:0];
    return n;
  }
}

int reticulum_ble_write(reticulum_ble_t h, const uint8_t* data, int data_len) {
  if (h == NULL) return -1;
  if (data == NULL || data_len <= 0) return 0;
  ReticulumBLE* obj = (__bridge ReticulumBLE*)h;
  int sent = 0;

  // Central mode: write to RX characteristic of connected peripheral.
  if (obj.connected && obj.peripheral != nil && obj.rxChar != nil) {
    CBCharacteristicWriteType wt = CBCharacteristicWriteWithoutResponse;
    if (!(obj.rxChar.properties & CBCharacteristicPropertyWriteWithoutResponse) && (obj.rxChar.properties & CBCharacteristicPropertyWrite)) {
      wt = CBCharacteristicWriteWithResponse;
    }
    NSInteger maxLen = [obj.peripheral maximumWriteValueLengthForType:wt];
    if (maxLen <= 0) maxLen = 20;

    while (sent < data_len) {
      int n = (int)MIN((NSInteger)(data_len - sent), maxLen);
      NSData* chunk = [NSData dataWithBytes:(data + sent) length:(NSUInteger)n];
      [obj.peripheral writeValue:chunk forCharacteristic:obj.rxChar type:wt];
      sent += n;
    }
  }

  // Peripheral mode: notify subscribed centrals via TX characteristic.
  if (obj.advertising && obj.peripheralMgr != nil && obj.periphTX != nil && obj.subscribers.count > 0) {
    @synchronized(obj.txPending) {
      [obj.txPending appendBytes:data length:(NSUInteger)data_len];
    }
    [obj flushPeripheralTX];
  }

  // If neither mode is active, still claim bytes accepted (like other transports buffering).
  if (sent == 0) {
    sent = data_len;
  }
  return sent;
}

const char* reticulum_ble_debug_string(reticulum_ble_t h) {
  if (h == NULL) return "ble://";
  ReticulumBLE* obj = (__bridge ReticulumBLE*)h;
  return obj.debugStr.UTF8String ?: "ble://";
}
