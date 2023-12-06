#include <HTTPClient.h>

#include <ESPmDNS.h>

#include <ESPConnect.h>
#include <espconnect_webpage.h>

#include <AsyncEventSource.h>
#include <AsyncJson.h>
#include <AsyncWebSocket.h>
#include <AsyncWebSynchronization.h>
#include <ESPAsyncWebServer.h>
#include <StringArray.h>
#include <WebAuthentication.h>
#include <WebHandlerImpl.h>
#include <WebResponseImpl.h>

#define CAMERA_MODEL_WROVER_KIT
// #define DING_DOORBELL // comment this out for camera-only build
#define DEBUG

#include "esp_camera.h"
#include "AsyncWebCamera.h"
#include "pins.h"

#ifdef DING_DOORBELL
const char* AP_ssid = "doorbell_dinghs_uuid";
#else if
const char* AP_ssid = "camera_dinghs_uuid";
#endif

const char* AP_password = "dinghomesecurity";

String baseIp;
int deviceId;
String notifyURL;

bool connectedToWifi = false;
bool connectedToBase = false;

int MOT_SENS_PIN = 14;
int DEBUG_LED_PIN = 2;

AsyncWebServer server(80);
HTTPClient http;

#ifdef DING_DOORBELL
#define IR_SENS_PIN 14
#endif

void initServer();
camera_config_t defaultCamCfg();

#include <esp_wifi.h>

void setup() {
#ifdef DING_DOORBELL
  pinMode(IR_SENS_PIN, INPUT);
  pinMode(2, OUTPUT);
#endif
  Serial.begin(115200);
  Serial.setDebugOutput(true);
  Serial.println();

  ESPConnect.autoConnect(AP_ssid);

  if (!ESPConnect.begin(&server)) {
    #ifdef DEBUG
      Serial.println("Failed to connect to WiFi, restarting...");
    #endif
    delay(5000);
    ESP.restart();
  }

  while (!WiFi.isConnected());

  #ifdef DEBUG
    Serial.println("Connected to WiFi"); 
  #endif

  connectedToWifi = true;

  // camera init
  camera_config_t config = defaultCamCfg();
  esp_err_t err = esp_camera_init(&config);
  if (err != ESP_OK) {
    #ifdef DEBUGs
      Serial.printf("Camera init failed with error 0ax%x, restarting...", err);
    #endif
    delay(5000);
    ESP.restart();
  }
  sensor_t* s = esp_camera_sensor_get();
  s->set_framesize(s, FRAMESIZE_VGA);

  if (!MDNS.begin(AP_ssid)) {
    #ifdef DEBUG
      Serial.println("Error setting up MDNS server, restarting...");
    #endif
    delay(5000);
    ESP.restart();
  }

  Serial.println("mDNS server started");
  MDNS.addService("dinghs", "tcp", 80);

  initServer();
  server.begin();

  Serial.println("Stream server started");

  #ifdef DING_DOORBELL
  pinMode(MOT_SENS_PIN, INPUT);
  pinMode(DEBUG_LED_PIN, OUTPUT);
  #endif
}

void loop(){
  if (!connectedToBase) return;
  #ifdef DING_DOORBELL
  if (digitalRead(MOT_SENS_PIN) == HIGH) {
    #ifdef DEBUG
    digitalWrite(DEBUG_LED_PIN, HIGH);
    #endif
    http.begin(notifyURL.c_str());
    int status = http.GET();

    if (status != 200) {
      #ifdef DEBUG
        Serial.println("Failed to notify base station");
      #endif
    }

    http.end();

    delay(3000);
  } else {
    digitalWrite(DEBUG_LED_PIN, LOW);
  }
  #endif
}


camera_config_t defaultCamCfg() {
  camera_config_t config;

  config.ledc_channel = LEDC_CHANNEL_0;
  config.ledc_timer = LEDC_TIMER_0;
  config.pin_d0 = Y2_GPIO_NUM;
  config.pin_d1 = Y3_GPIO_NUM;
  config.pin_d2 = Y4_GPIO_NUM;
  config.pin_d3 = Y5_GPIO_NUM;
  config.pin_d4 = Y6_GPIO_NUM;
  config.pin_d5 = Y7_GPIO_NUM;
  config.pin_d6 = Y8_GPIO_NUM;
  config.pin_d7 = Y9_GPIO_NUM;
  config.pin_xclk = XCLK_GPIO_NUM;
  config.pin_pclk = PCLK_GPIO_NUM;
  config.pin_vsync = VSYNC_GPIO_NUM;
  config.pin_href = HREF_GPIO_NUM;
  config.pin_sscb_sda = SIOD_GPIO_NUM;
  config.pin_sscb_scl = SIOC_GPIO_NUM;
  config.pin_pwdn = PWDN_GPIO_NUM;
  config.pin_reset = RESET_GPIO_NUM;
  config.xclk_freq_hz = 20000000; 
  config.pixel_format = PIXFORMAT_JPEG;

  if (psramFound()) {
    config.frame_size = FRAMESIZE_UXGA;
    config.jpeg_quality = 10;
    config.fb_count = 2;
  } else {
    config.frame_size = FRAMESIZE_SVGA;
    config.jpeg_quality = 12;
    config.fb_count = 1;
  }

  return config;
}

void linkBase(AsyncWebServerRequest *request) {
  if (connectedToBase){ 
    request->send(200, "application/json", "{\"message\": \"device already connected to a base station\"}");
    return;
  }

  int numParams = request->params();

  if (numParams < 2) {
    request->send(400, "application/json", "{\"error\": \"invalid url parameters\"}");
    return;
  }

  AsyncWebParameter* id = request->getParam(0);
  AsyncWebParameter* ip = request->getParam(1);

  if (id->name() != "id") {
    request->send(400, "application/json", "{\"error\": \"invalid url parameters\"}");
    return;
  }

  deviceId = id->value().toInt();

  if (ip->name() != "ip") {
    request->send(400, "application/json", "{\"error\": \"invalid url parameters\"}");
    return;
  }

  baseIp = ip->value();

  notifyURL = "http://" + baseIp + ":8080/api/notify/" + deviceId;

  #ifdef DEBUG
  Serial.println("Base station notification URL: " + notifyURL);
  #endif

  connectedToBase = true;

  request->send(200, "application/json", "{\"message\": \"connected to base station\"}");
}

void initServer() {
  server.on("/capture", HTTP_GET, sendJpg);
  server.on("/stream", HTTP_GET, streamJpg);
  server.on("/link", HTTP_GET, linkBase);
}
