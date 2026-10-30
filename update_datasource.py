path = '/home/afari/Projects/ns/frontend/nsapp/lib/features/seeker/data/datasource/remote/seeker_remote_datasource_impl.dart'
with open(path, 'r') as f:
    content = f.read()

old_code = '''  @override
  Future<bool> updateRequest({required Request request, String? imagePath}) async {
    try {
      final token = await Helpers.getString("token");
      final updateData = request.toUpdateJson();
     
      final response = await _dio.patch(
        "$baseRequestUrl/services/requests/${request.id}/",
        data: updateData,
        options: Options(headers: dioHeaders(token)),
      );
      if (response.statusCode == 200) {
        if ((request.withImage ?? false) && imagePath != null) {
          // 1. Direct resilient upload to Prolead GoStore CAS (or 0ms instant cache hit)
          try {
            final proleadUrl = await MediaUtils.uploadRequest(File(imagePath));
            if (proleadUrl != null && proleadUrl.isNotEmpty) {
              final res = await _dio.patch(
                "$baseUrl/services/requests/image/",
                data: {"id": request.id, "image": proleadUrl},
                options: Options(headers: dioHeaders(token)),
              );
              if (res.statusCode == 200) {
                return true;
              }
            }
          } catch (e) {
        // Error intentionally suppressed
      }

          // 2. Fallback to multipart
          final res = await _dio.patch(
            "$baseUrl/services/requests/image/",
            data: FormData.fromMap({
              "data": json.encode({"id": request.id}),
              "image": await MultipartFile.fromFile(imagePath),
            }),
            options: Options(headers: dioMultiPartHeaders(token)),
          );
          if (res.statusCode == 200) {
            return true;
          } else {
            return true;
          }
        }
        return true;
      }
      throw Exception('Failed');
    } catch (e) { rethrow; }'''

new_code = '''  @override
  Future<bool> updateRequest({required Request request, String? imagePath, String? videoPath}) async {
    try {
      final token = await Helpers.getString("token");
      final updateData = request.toUpdateJson();
     
      final response = await _dio.patch(
        "$baseRequestUrl/services/requests/${request.id}/",
        data: updateData,
        options: Options(headers: dioHeaders(token)),
      );
      if (response.statusCode == 200) {
        if ((request.withImage ?? false) && imagePath != null) {
          // 1. Direct resilient upload to Prolead GoStore CAS (or 0ms instant cache hit)
          try {
            final proleadUrl = await MediaUtils.uploadRequest(File(imagePath));
            if (proleadUrl != null && proleadUrl.isNotEmpty) {
              await _dio.patch(
                "$baseUrl/services/requests/image/",
                data: {"id": request.id, "image": proleadUrl},
                options: Options(headers: dioHeaders(token)),
              );
            }
          } catch (e) {
            // Error intentionally suppressed
          }

          // 2. Fallback to multipart
          try {
            await _dio.patch(
              "$baseUrl/services/requests/image/",
              data: FormData.fromMap({
                "data": json.encode({"id": request.id}),
                "image": await MultipartFile.fromFile(imagePath),
              }),
              options: Options(headers: dioMultiPartHeaders(token)),
            );
          } catch (_) {}
        }

        // Upload/Update Job Video to GoStore if present
        if (videoPath != null && videoPath.isNotEmpty) {
          try {
            final videoFile = File(videoPath);
            if (await videoFile.exists()) {
              final uploadedUrl = await MediaUtils.uploadToGoStore(
                file: videoFile,
                subfolder: 'request_videos',
              );
              if (uploadedUrl != null && uploadedUrl.isNotEmpty) {
                await _dio.patch(
                  "$baseUrl/services/requests/video/",
                  data: {"id": request.id, "video": uploadedUrl},
                  options: Options(headers: dioHeaders(token)),
                );
              } else {
                await _dio.patch(
                  "$baseUrl/services/requests/video/",
                  data: FormData.fromMap({
                    "data": json.encode({"id": request.id}),
                    "video": await MultipartFile.fromFile(videoPath),
                  }),
                  options: Options(headers: dioMultiPartHeaders(token)),
                );
              }
            }
          } catch (_) {
            try {
              await _dio.patch(
                "$baseUrl/services/requests/video/",
                data: FormData.fromMap({
                  "data": json.encode({"id": request.id}),
                  "video": await MultipartFile.fromFile(videoPath),
                }),
                options: Options(headers: dioMultiPartHeaders(token)),
              );
            } catch (_) {}
          }
        }

        return true;
      }
      throw Exception('Failed');
    } catch (e) { rethrow; }'''

assert old_code in content, 'old_code not found in seeker_remote_datasource_impl.dart'
content = content.replace(old_code, new_code)

with open(path, 'w') as f:
    f.write(content)
print('Updated updateRequest in seeker_remote_datasource_impl.dart')

