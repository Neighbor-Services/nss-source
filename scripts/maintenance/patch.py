import re

f = open('e:/ns/frontend/nsapp/lib/features/shared/presentation/widget/appointment_detail_bottom_sheet.dart', 'r', encoding='utf-8')
c = f.read()
f.close()

verify_code_func = """
  Future<void> _verifyCode() async {
    final appt = widget.data.appointment;
    if (appt == null || _codeController.text.trim().isEmpty) return;
    setState(() => _isVerifying = true);
    try {
      final token = await Helpers.getString("token");
      final response = await dio.post(
        "$baseUrl/interactions/appointments/${appt.id}/verify-code/",
        data: {'code': _codeController.text.trim()},
        options: Options(headers: dioHeaders(token)),
      );
      if (response.statusCode == 200) {
        setState(() {
           appt.status = 'IN_PROGRESS';
        });
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text("Code verified successfully!")),
          );
        }
      }
    } on DioException catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.response?.data['error'] ?? "Failed to verify code")),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text("Error verifying code")),
        );
      }
    } finally {
      if (mounted) setState(() => _isVerifying = false);
    }
  }
"""

if "_verifyCode" not in c:
    c = c.replace('  @override\n  Widget build(BuildContext context) {', verify_code_func + '\n  @override\n  Widget build(BuildContext context) {')

ui_section = """            if (!_isEditing && widget.data.role == 'seeker' && appt.secretCode != null) ...[
              Container(
                width: double.infinity,
                padding: EdgeInsets.symmetric(vertical: 20.h, horizontal: 16.w),
                decoration: BoxDecoration(
                  color: context.appColors.primaryColor.withAlpha(20),
                  borderRadius: BorderRadius.circular(16.r),
                  border: Border.all(color: context.appColors.primaryColor.withAlpha(50), width: 1.5.r),
                ),
                child: Column(
                  children: [
                    Text(
                      "VERIFICATION CODE",
                      style: TextStyle(
                        fontSize: 10.sp,
                        fontWeight: FontWeight.bold,
                        color: context.appColors.primaryColor,
                        letterSpacing: 1.5,
                      ),
                    ),
                    SizedBox(height: 8.h),
                    Text(
                      appt.secretCode!,
                      style: TextStyle(
                        fontSize: 32.sp,
                        fontWeight: FontWeight.w900,
                        letterSpacing: 8.0,
                        color: contentColor,
                      ),
                    ),
                    SizedBox(height: 6.h),
                    Text(
                      "Show this code to the provider in person",
                      style: TextStyle(
                        fontSize: 12.sp,
                        color: secondaryTextColor,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
              ),
              SizedBox(height: 32.h),
            ],

            if (!_isEditing && widget.data.role == 'provider' && appt.status == 'SCHEDULED') ...[
              Container(
                width: double.infinity,
                padding: EdgeInsets.symmetric(vertical: 20.h, horizontal: 16.w),
                decoration: BoxDecoration(
                  color: context.appColors.secondaryColor.withAlpha(20),
                  borderRadius: BorderRadius.circular(16.r),
                  border: Border.all(color: context.appColors.secondaryColor.withAlpha(50), width: 1.5.r),
                ),
                child: Column(
                  children: [
                    Text(
                      "VERIFY START",
                      style: TextStyle(
                        fontSize: 10.sp,
                        fontWeight: FontWeight.bold,
                        color: context.appColors.secondaryColor,
                        letterSpacing: 1.5,
                      ),
                    ),
                    SizedBox(height: 12.h),
                    TextField(
                      controller: _codeController,
                      textAlign: TextAlign.center,
                      maxLength: 6,
                      textCapitalization: TextCapitalization.characters,
                      style: TextStyle(
                        fontSize: 24.sp,
                        fontWeight: FontWeight.bold,
                        letterSpacing: 6.0,
                        color: contentColor,
                      ),
                      decoration: InputDecoration(
                        counterText: "",
                        hintText: "ENTER CODE",
                        hintStyle: TextStyle(
                           fontSize: 16.sp,
                           letterSpacing: 1.0,
                           color: secondaryTextColor,
                        ),
                        filled: true,
                        fillColor: context.appColors.cardBackground,
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(12.r),
                          borderSide: BorderSide.none,
                        ),
                        enabledBorder: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(12.r),
                          borderSide: BorderSide.none,
                        ),
                        focusedBorder: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(12.r),
                          borderSide: BorderSide.none,
                        ),
                      ),
                    ),
                    SizedBox(height: 12.h),
                    _isVerifying 
                      ? const CircularProgressIndicator() 
                      : SizedBox(
                          width: double.infinity,
                          child: SolidButton(
                            label: "VERIFY ARRIVAL",
                            onPressed: _verifyCode,
                            color: context.appColors.secondaryColor,
                            height: 48.h,
                            textColor: Colors.white,
                          ),
                        ),
                  ],
                ),
              ),
              SizedBox(height: 32.h),
            ],

            if (!_isEditing && widget.data.role == 'provider' && appt.status == 'IN_PROGRESS') ...[
               Container(
                width: double.infinity,
                padding: EdgeInsets.symmetric(vertical: 16.h, horizontal: 16.w),
                decoration: BoxDecoration(
                  color: Colors.green.withAlpha(20),
                  borderRadius: BorderRadius.circular(16.r),
                  border: Border.all(color: Colors.green.withAlpha(50), width: 1.5.r),
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(FontAwesomeIcons.circleCheck, color: Colors.green, size: 20.r),
                    SizedBox(width: 8.w),
                    Text(
                      "SERVICE VERIFIED",
                      style: TextStyle(
                        fontSize: 12.sp,
                        fontWeight: FontWeight.bold,
                        color: Colors.green,
                        letterSpacing: 0.5,
                      ),
                    ),
                  ],
                ),
              ),
              SizedBox(height: 32.h),
            ],"""

c = re.sub(r'if \(!_isEditing && appt\.secretCode != null\) \.\.\.\[[\s\S]*?SizedBox\(height: 32\.h\),\n            \],', ui_section, c)

f2 = open('e:/ns/frontend/nsapp/lib/features/shared/presentation/widget/appointment_detail_bottom_sheet.dart', 'w', encoding='utf-8')
f2.write(c)
f2.close()
print("PATCH SUCCESS")
