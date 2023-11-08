#include <SPIFFS.h>

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

AsyncWebServer apServer(80);
AsyncWebServer streamServer(80);

extern bool connectedToWifi;

void startApServer() 
{ 
  apServer.serveStatic("/", SPIFFS, "/res");

  apServer.on("/", HTTP_GET, [](AsyncWebServerRequest* req) {
    req->send(SPIFFS, "/ap.html");
  });

  apServer.on("/connect", HTTP_POST, [connectedToWifi](AsyncWebServerRequest* req) {
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

    connectedToWifi = true;

    req->send(200);
  });

  apServer.begin();
}

void stopApServer() 
{
  apServer.end();
}

void startCameraServer() 
{
  streamServer.on("/capture", HTTP_GET, sendJpg);
  streamServer.on("/stream", HTTP_GET, streamJpg);

  streamServer.begin();
}

void stopCameraServer()
{
  streamServer.end();
}
