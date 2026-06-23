module.exports = function (api) {
  api.cache(true);
  return {
    presets: [
      ["babel-preset-expo", { jsxImportSource: "nativewind" }],
      "nativewind/babel",
    ],
    plugins: [
      // "react-native-reanimated/plugin", // Deshabilitado temporalmente (requiere react-native-worklets-core en v4)
    ],
  };
};
