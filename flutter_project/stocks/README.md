# Stocks

[Example from the flutter repo](https://github.com/flutter/flutter/tree/master/examples/stocks)  
Demo app for the material design widgets and other features provided by Flutter.

# Difference

The only edit made was to change the font to Roboto.  
_If the host is missing some fonts, it can cause the text to not be rendered or
worse the app might crash_

```diff
diff --git a/examples/stocks/lib/main.dart b/examples/stocks/lib/main.dart
index d415902d7..eba3362d9 100644
--- a/examples/stocks/lib/main.dart
+++ b/examples/stocks/lib/main.dart
@@ -71,12 +71,14 @@ class StocksAppState extends State<StocksApp> {
       case StockMode.optimistic:
         return new ThemeData(
           brightness: Brightness.light,
-          primarySwatch: Colors.purple
+          primarySwatch: Colors.purple,
+          fontFamily: 'Roboto'
         );
       case StockMode.pessimistic:
         return new ThemeData(
           brightness: Brightness.dark,
-          accentColor: Colors.redAccent
+          accentColor: Colors.redAccent,
+          fontFamily: 'Roboto'
         );
     }
     assert(_configuration.stockMode != null);
diff --git a/examples/stocks/pubspec.yaml b/examples/stocks/pubspec.yaml
index 162a7914e..6dba3fbca 100644
--- a/examples/stocks/pubspec.yaml
+++ b/examples/stocks/pubspec.yaml
@@ -74,4 +74,18 @@ dev_dependencies:
 flutter:
   uses-material-design: true
 
-# PUBSPEC CHECKSUM: cf1e
+  fonts:
+    - family: Roboto
+      fonts:
+        - asset: fonts/Roboto/Roboto-Thin.ttf
+          weight: 100
+        - asset: fonts/Roboto/Roboto-Light.ttf
+          weight: 300
+        - asset: fonts/Roboto/Roboto-Regular.ttf
+          weight: 400
+        - asset: fonts/Roboto/Roboto-Medium.ttf
+          weight: 500
+        - asset: fonts/Roboto/Roboto-Bold.ttf
+          weight: 700
+        - asset: fonts/Roboto/Roboto-Black.ttf
+          weight: 900
```

