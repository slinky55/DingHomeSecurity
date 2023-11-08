#include <DNSServer.h>
#include <SPIFFS.h>
#include "esp_camera.h"
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
#include <WiFi.h>
#include "AsyncWebCamera.h"

#define CAMERA_MODEL_WROVER_KIT 

#include "pins.h"

const char* AP_ssid = "cam_module_id";  // or doorbell_module_id
const char* AP_password = "dinghomesecurity";

AsyncWebServer server(80);

void startServer();

void setup() {
  Serial.begin(115200);
  Serial.setDebugOutput(true);
  Serial.println();

  if(!SPIFFS.begin(true)){
    Serial.println("An Error has occurred while mounting SPIFFS");
    return;
  }

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

  // camera init
  esp_err_t err = esp_camera_init(&config);
  if (err != ESP_OK) {
    Serial.printf("Camera init failed with error 0ax%x", err);
    return;
  }

  sensor_t * s = esp_camera_sensor_get();
  // drop down frame size for higher initial frame rate
  s->set_framesize(s, FRAMESIZE_VGA);

  WiFi.disconnect();
  WiFi.mode(WIFI_AP);
  Serial.println("Setting soft-AP ... ");
  boolean result = WiFi.softAP(AP_ssid, AP_password);
  if (result) {
    Serial.println("Ready");
    Serial.println(String("Soft-AP IP address = ") + WiFi.softAPIP().toString());
    Serial.println(String("MAC address = ") + WiFi.softAPmacAddress().c_str());
  } else {
    Serial.println("Failed!");
  }

  startServer();
}

void loop() {
  delay(10000);
}

void startServer() 
{ 
  server.serveStatic("/", SPIFFS, "/res");

  server.on("/", HTTP_GET, [](AsyncWebServerRequest* req) {
    req->send(SPIFFS, "/ap.html");
  });

  server.on("/connect", HTTP_POST, [](AsyncWebServerRequest* req) {
    String s;
    String p;

    if (req->hasParam("ssid", true)) {
      s = req->getParam("ssid", true)->value();
    } else {
      req->send(500, "text/plain", "Couldn't find SSID param");
      return;
    }

    if (req->hasParam("password", true)) {
      p = req->getParam("password", true)->value();
    } else {
      req->send(500, "text/plain", "Couldn't find SSID param");
      return;
    }

    WiFi.disconnect();
    WiFi.mode(WIFI_STA);
    WiFi.begin(s, p);
    
    while (WiFi.status() != WL_CONNECTED) {
      Serial.print(".");
      delay(500);
    }

    Serial.println();
    Serial.println("Connected to wifi network");

    req->send(200);
  });

  server.on("/capture", HTTP_GET, sendJpg);
  server.on("/stream", HTTP_GET, streamJpg);

  server.begin();
}
