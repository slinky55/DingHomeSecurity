#include <ESPmDNS.h>

#include <ESPConnect.h>
#include <espconnect_webpage.h>

#include <AsyncEventSource.h>
#include <AsyncJson.h>
#include <AsyncWebSocket.h>
#include <AsyncWebSynchronization.h>
#include <ESPAsyncWebServer.h>
#include <SPIFFSEditor.h>
#include <StringArray.h>
#include <WebAuthentication.h>
#include <WebHandlerImpl.h>
#include <WebResponseImpl.h>

#define CAMERA_MODEL_WROVER_KIT 
//#define DING_DOORBELL // comment this out for camera-only build

#include "esp_camera.h"
#include "AsyncWebCamera.h"
#include "pins.h"

#ifdef DING_DOORBELL
const char* AP_ssid = "doorbell_module_id";
#else if
const char* AP_ssid = "camera_module_id";
#endif

const char* AP_password = "dinghomesecurity";

bool connectedToWifi = false;
bool connectedToBase = false;

AsyncWebServer server(80);

#ifdef DING_DOORBELL
#define IR_SENS_PIN 14
#endif 

void initServer();
camera_config_t defaultCamCfg();

void setup() {
  #ifdef DING_DOORBELL
    pinMode(IR_SENS_PIN, INPUT);
    pinMode(2, OUTPUT);
  #endif
  
  Serial.begin(115200);
  Serial.setDebugOutput(true);
  Serial.println();

  ESPConnect.autoConnect(AP_ssid);

  if(ESPConnect.begin(&server)){
    Serial.println("Connected to WiFi");
    Serial.println("IPAddress: " + WiFi.localIP().toString());
  }else{
    Serial.println("Failed to connect to WiFi");
  }

  // camera init
  camera_config_t config = defaultCamCfg();
  esp_err_t err = esp_camera_init(&config);
  if (err != ESP_OK) {
    Serial.printf("Camera init failed with error 0ax%x", err);
    return;
  }
  sensor_t * s = esp_camera_sensor_get();
  s->set_framesize(s, FRAMESIZE_VGA);

  if (!MDNS.begin("esp32")) {
      Serial.println("Error setting up MDNS server!");
      while(1) {
          delay(1000);
      }
  }
  Serial.println("mDNS server started");

  initServer();
  server.begin();

  MDNS.addService("http", "tcp", 80);
}

void loop() {
  #ifdef DING_DOORBELL
    // TODO: Check for doorbell stuff and send to base
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

  if(psramFound()){
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

void initServer() {
  server.on("/capture", HTTP_GET, sendJpg);
  server.on("/stream", HTTP_GET, streamJpg);
}
