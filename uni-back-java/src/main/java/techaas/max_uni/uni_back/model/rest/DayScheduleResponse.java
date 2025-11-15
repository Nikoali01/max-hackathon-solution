package techaas.max_uni.uni_back.model.rest;

import com.fasterxml.jackson.annotation.JsonFormat;
import lombok.AllArgsConstructor;
import lombok.Data;

import java.time.LocalTime;

@Data
@AllArgsConstructor
public class DayScheduleResponse {

    @JsonFormat(pattern = "HH:mm")
    private LocalTime time;
    private String discipline;
    private String instructor;
    private String location;
    private String description;
}
