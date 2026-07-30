import "./global.css";

import { StatusBar } from "expo-status-bar";
import { Text, View } from "react-native";


export default function App() {
  return (
    <View className="flex-1 items-center justify-center bg-white px-6">
      <Text className="text-3xl font-bold text-slate-900">LetsInit Expo</Text>
      <Text className="mt-2 text-center text-base text-slate-600">
        Expo SDK 57 with NativeWind styling is ready.
      </Text>
      <StatusBar style="auto" />
    </View>
  );
}
