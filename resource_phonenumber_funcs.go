package main

import (
	"fmt"

	twilioc "github.com/f0rk/twiliogo"
	"github.com/hashicorp/terraform/helper/schema"
)

func phonenumberCreate(d *schema.ResourceData, meta interface{}) error {
	m := meta.(*twilioMeta)

	var filters []twilioc.Optional

	if v, ok := d.GetOk("location"); ok {
		locationFilter := (v.(*schema.Set)).List()[0].(map[string]interface{})
		for k, v := range locationFilter {
			switch k {
			case "near_number":
				filters = append(filters, twilioc.NearNumber(v.(string)))
			case "near_lat_long":
				if (v.(*schema.Set)).Len() != 0 {
					latLong := (v.(*schema.Set)).List()[0].(map[string]interface{})

					filters = append(filters, twilioc.NearLatLong{
						Latitude:  latLong["latitude"].(float64),
						Longitude: latLong["longitude"].(float64),
					})
				}
			case "distance":
				filters = append(filters, twilioc.Distance(v.(int)))
			case "postal_code":
				filters = append(filters, twilioc.InPostalCode(v.(string)))
			case "rate_center":
				filters = append(filters, twilioc.InRateCenter(v.(string)))
			case "lata":
				filters = append(filters, twilioc.InLata(v.(string)))
			}
		}
	}

	numbers, err := twilioc.GetLocalAvailablePhoneNumbers(
		m.Client,
		d.Get("iso_country_code").(string),
		filters...,
	)

	if err != nil {
		return err
	}

	if len(*numbers) == 0 {
		return fmt.Errorf("Twilio API returned no numbers matching criteria")
	}

	numberStruct, err := twilioc.BuyPhoneNumber(
		m.Client,
		twilioc.PhoneNumber((*numbers)[0].PhoneNumber),
	)
	if err != nil {
		return err
	}

	d.SetId(numberStruct.Sid)

	return phonenumberUpdate(d, meta)
}

func phonenumberRead(d *schema.ResourceData, meta interface{}) error {
	m := meta.(*twilioMeta)

	numberStruct, err := twilioc.GetIncomingPhoneNumber(m.Client, d.Id())
	if err != nil {
		return err
	}

	if numberStruct == nil {
		d.SetId("")
		return err
	}

	d.Set("name", numberStruct.FriendlyName)
	d.Set("phone_number", numberStruct.PhoneNumber)
	d.Set("voice_url", numberStruct.VoiceUrl)
	d.Set("voice_method", numberStruct.VoiceMethod)
	d.Set("voice_fallback_url", numberStruct.VoiceFallbackUrl)
	d.Set("voice_fallback_method", numberStruct.VoiceFallbackMethod)
	d.Set("status_callback", numberStruct.StatusCallback)
	d.Set("status_callback_method", numberStruct.StatusCallbackMethod)
	d.Set("voice_caller_id_lookup", numberStruct.VoiceCallerIdLookup)
	d.Set("voice_application_sid", numberStruct.VoiceApplicationSid)
	d.Set("date_created", numberStruct.DateCreated)
	d.Set("date_updated", numberStruct.DateUpdated)
	d.Set("sms_url", numberStruct.SmsUrl)
	d.Set("sms_method", numberStruct.SmsMethod)
	d.Set("sms_fallback_url", numberStruct.SmsFallbackUrl)
	d.Set("sms_fallback_method", numberStruct.SmsFallbackMethod)
	d.Set("sms_application_sid", numberStruct.SmsApplicationSid)
	//d.Set("capabilities", Capabilities         Capabilites `json:"capabilities"`
	d.Set("api_version", numberStruct.ApiVersion)
	d.Set("uri", numberStruct.Uri)

	return nil
}

func phonenumberUpdate(d *schema.ResourceData, meta interface{}) error {
	m := meta.(*twilioMeta)

	var voiceCallerIDLookup *bool

	incomingNumber := new(twilioc.IncomingPhoneNumber)
	incomingNumber.Sid = d.Id()

	if d.HasChange("name") {
		incomingNumber.FriendlyName = d.Get("name").(string)
	}
	if d.HasChange("api_version") {
		incomingNumber.ApiVersion = d.Get("api_version").(string)
	}
	if d.HasChange("voice_caller_id_lookup") {
		voiceCallerIDLookup = new(bool)
		*voiceCallerIDLookup = d.Get("voice_caller_id_lookup").(bool)
	}
	if d.HasChange("voice_url") {
		incomingNumber.VoiceUrl = d.Get("voice_url").(string)
	}
	if d.HasChange("voice_method") {
		incomingNumber.VoiceMethod = d.Get("voice_method").(string)
	}
	if d.HasChange("voice_fallback_url") {
		incomingNumber.VoiceFallbackUrl = d.Get("voice_fallback_url").(string)
	}
	if d.HasChange("voice_fallback_method") {
		incomingNumber.VoiceFallbackMethod = d.Get("voice_fallback_method").(string)
	}
	if d.HasChange("voice_application_sid") {
		incomingNumber.VoiceApplicationSid = d.Get("voice_application_sid").(string)
	}
	if d.HasChange("sms_url") {
		incomingNumber.SmsUrl = d.Get("sms_url").(string)
	}
	if d.HasChange("sms_method") {
		incomingNumber.SmsMethod = d.Get("sms_method").(string)
	}
	if d.HasChange("sms_fallback_url") {
		incomingNumber.SmsFallbackUrl = d.Get("sms_fallback_url").(string)
	}
	if d.HasChange("sms_fallback_method") {
		incomingNumber.SmsFallbackMethod = d.Get("sms_fallback_method").(string)
	}
	if d.HasChange("sms_application_sid") {
		incomingNumber.SmsApplicationSid = d.Get("sms_application_sid").(string)
	}
	if d.HasChange("status_callback") {
		incomingNumber.StatusCallback = d.Get("status_callback").(string)
	}
	if d.HasChange("status_callback_method") {
		incomingNumber.StatusCallbackMethod = d.Get("status_callback_method").(string)
	}

	_, err := twilioc.UpdatePhoneNumberFields(m.Client, incomingNumber, voiceCallerIDLookup)
	if err != nil {
		return err
	}

	return phonenumberRead(d, meta)
}

func phonenumberDelete(d *schema.ResourceData, meta interface{}) error {
	m := meta.(*twilioMeta)

	err := twilioc.ReleasePhoneNumber(m.Client, d.Id())

	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
